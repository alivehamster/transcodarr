package libs

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
)

func NewJobScheduler() *JobScheduler {
	return &JobScheduler{
		scheduler: cron.New(),
		jobMap:    make(map[int]cron.EntryID),
	}
}

func (js *JobScheduler) StartJobs(db *sql.DB) error {
	rows, err := db.Query("SELECT id, cron FROM libraries")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var lib Library
		if err := rows.Scan(&lib.ID, &lib.Cron); err != nil {
			return err
		}
		id, _ := js.scheduler.AddFunc(lib.Cron, func() {
			job(db, lib.ID)
		})
		js.jobMap[lib.ID] = id
	}
	if err := rows.Err(); err != nil {
		return err
	}

	js.scheduler.Start()
	return nil
}

func (js *JobScheduler) EditSchedule(db *sql.DB, lib Library) error {
	js.mu.Lock()
	defer js.mu.Unlock()

	if entryID, exists := js.jobMap[lib.ID]; exists {
		js.scheduler.Remove(entryID)
	}

	id, err := js.scheduler.AddFunc(lib.Cron, func() {
		job(db, lib.ID)
	})
	if err != nil {
		return err
	}

	js.jobMap[lib.ID] = id
	return nil
}

func (js *JobScheduler) DeleteJob(libID int) {
	js.mu.Lock()
	defer js.mu.Unlock()

	if entryID, exists := js.jobMap[libID]; exists {
		js.scheduler.Remove(entryID)
		delete(js.jobMap, libID)
	}
}

func job(db *sql.DB, id int) {
	var lib Library
	var skiplist []Skip
	var skiplistJSON sql.NullString
	var configJSON string
	var history []string
	defer func() { SaveHistoryBatch(db, history) }()

	row := db.QueryRow("SELECT id, name, cron, config, skiplist FROM libraries WHERE id = ?", id)
	err := row.Scan(&lib.ID, &lib.Name, &lib.Cron, &configJSON, &skiplistJSON)
	if err != nil {
		println("Failed to get library:", err.Error())
		return
	}

	println("Running job for library:", lib.Name)

	if err := json.Unmarshal([]byte(configJSON), &lib.Config); err != nil {
		println("Failed to parse library config:", err.Error())
		return
	}

	if skiplistJSON.Valid && skiplistJSON.String != "" {
		if err := json.Unmarshal([]byte(skiplistJSON.String), &skiplist); err != nil {
			println("Failed to parse library skiplist:", err.Error())
			return
		}
	}

	skipMap := make(map[string]struct{}, len(skiplist))
	for _, s := range skiplist {
		skipMap[s.Path] = struct{}{}
	}

	files := getlibItems(lib)
	for _, path := range files {
		if _, shouldSkip := skipMap[path]; shouldSkip {
			msg := fmt.Sprintf("Skipping: %s", path)
			println(msg)
			history = append(history, msg)
			continue
		}

		info, err := os.Stat(path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		// Cast the Sys() interface to the platform-specific Stat_t type
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			fmt.Println("Not a Unix-like system; cannot check hardlinks and ctime")
			continue
		}

		if stat.Nlink > 1 {
			msg := fmt.Sprintf("Skipping file with multiple hardlinks: %s", path)
			println(msg)
			history = append(history, msg)
			continue
		}
		if time.Since(info.ModTime()) < (30 * 24 * time.Hour) {
			msg := fmt.Sprintf("Skipping recently changed file: %s", path)
			println(msg)
			history = append(history, msg)
			continue
		}

		codec, err := getCodec(path)
		if err != nil {
			println("Failed to get codec for", path, ":", err.Error())
			continue
		}
		if codec == "av1" {
			msg := fmt.Sprintf("Skipping AV1 file: %s", path)
			println(msg)
			history = append(history, msg)

			skiplist, err = updateSkiplist(db, id, skiplist, Skip{Path: path, Description: fmt.Sprintf("Codec is already %s", codec)})
			if err != nil {
				println("Failed to update skiplist:", err.Error())
			}
			continue
		}
		msg := fmt.Sprintf("Processing: %s", path)
		println(msg)
		history = append(history, msg)

		// after transcoding compare new to initial file size
		filename := filepath.Base(path)
		dir := filepath.Dir(path)
		ext := filepath.Ext(filename)
		nameWithoutExt := strings.TrimSuffix(filename, ext)
		outputPath := filepath.Join(dir, nameWithoutExt+".tmp"+ext)

		if err := transcode(lib.Config, path, outputPath); err != nil {
			msg = fmt.Sprintf("Failed to transcode: %s", err.Error())
			println(msg)
			history = append(history, msg)
			os.Remove(outputPath)
			continue
		}

		outputInfo, err := os.Stat(outputPath)
		if err != nil {
			println("Failed to get output file info:", err.Error())
			continue
		}
		if outputInfo.Size() >= info.Size() {
			msg = fmt.Sprintf("Transcoded file is not smaller, skipping replacement: %s", path)
			println(msg)
			history = append(history, msg)
			os.Remove(outputPath)
			skiplist, err = updateSkiplist(db, id, skiplist, Skip{Path: path, Description: "transcoded file not smaller"})
			if err != nil {
				println("Failed to update skiplist:", err.Error())
			}
			continue
		}

		if err := os.Remove(path); err != nil {
			println("Failed to remove original file:", err.Error())
			continue
		}

		if err := os.Rename(outputPath, path); err != nil {
			println("Failed to rename transcoded file:", err.Error())
			continue
		}

		skiplist, err = updateSkiplist(db, id, skiplist, Skip{Path: path, Description: "successfully transcoded"})
		if err != nil {
			println("Failed to update skiplist:", err.Error())
		}
	}

}

func RunJob(db *sql.DB, id int) {
	job(db, id)
}
