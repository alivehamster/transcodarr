package libs

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
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
			js.runJob(db, lib.ID)
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
		js.runJob(db, lib.ID)
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

func (js *JobScheduler) runJob(db *sql.DB, id int) {
	js.jobMu.Lock()
	defer js.jobMu.Unlock()
	job(db, id)
}

func job(db *sql.DB, id int) {
	var lib Library
	var configJSON string
	row := db.QueryRow("SELECT id, name, cron, config FROM libraries WHERE id = ?", id)
	err := row.Scan(&lib.ID, &lib.Name, &lib.Cron, &configJSON)
	if err != nil {
		log.Printf("Failed to get library: %s", err.Error())
		return
	}

	log.Printf("Running job for library: %s", lib.Name)

	if err := json.Unmarshal([]byte(configJSON), &lib.Config); err != nil {
		log.Printf("Failed to parse library config: %s", err.Error())
		return
	}

	skipMap, err := getSkipMap(db, id)
	if err != nil {
		log.Printf("Failed to get skiplist: %s", err.Error())
		return
	}

	files := getlibItems(lib)
	for _, path := range files {
		if _, shouldSkip := skipMap[path]; shouldSkip {
			SaveHistory(db, logMsg(fmt.Sprintf("Skipping: %s", path)))
			continue
		}

		info, err := os.Stat(path)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}

		if lib.Config.MinimumFileSizeMb > 0 {
			minimumBytes := lib.Config.MinimumFileSizeMb * 1024 * 1024
			if info.Size() < minimumBytes {
				SaveHistory(db, logMsg(fmt.Sprintf("Skipping file smaller than minimum size (%d MB): %s", lib.Config.MinimumFileSizeMb, path)))
				err = addSkip(db, id, path, fmt.Sprintf("File is smaller than minimum size (%d MB)", lib.Config.MinimumFileSizeMb))
				if err != nil {
					log.Printf("Failed to add to skiplist: %s", err.Error())
				}
				continue
			}
		}

		if lib.Config.Hardlinks {
			// Cast the Sys() interface to the platform-specific Stat_t type
			stat, ok := info.Sys().(*syscall.Stat_t)
			if !ok {
				log.Println("Not a Unix-like system; cannot check hardlinks and ctime")
				continue
			}

			if stat.Nlink > 1 {
				SaveHistory(db, logMsg(fmt.Sprintf("Skipping file with multiple hardlinks: %s", path)))
				continue
			}
		}

		if lib.Config.FileAge != 0 {
			if time.Since(info.ModTime()) < (time.Duration(lib.Config.FileAge) * 24 * time.Hour) {
				SaveHistory(db, logMsg(fmt.Sprintf("Skipping recently changed file: %s", path)))
				continue
			}
		}

		if len(lib.Config.MediaCodec) > 0 {
			codec, err := getCodec(path)
			if err != nil {
				log.Printf("Failed to get codec for %s: %s", path, err.Error())
				continue
			}
			if slices.Contains(lib.Config.MediaCodec, codec) {
				SaveHistory(db, logMsg(fmt.Sprintf("Skipping file with codec %s: %s", codec, path)))

				err = addSkip(db, id, path, fmt.Sprintf("Codec is already %s", codec))
				if err != nil {
					log.Printf("Failed to add to skiplist: %s", err.Error())
				}
				continue
			}
		}

		SaveHistory(db, logMsg(fmt.Sprintf("Processing: %s", path)))

		// after transcoding compare new to initial file size
		filename := filepath.Base(path)
		dir := filepath.Dir(path)
		ext := filepath.Ext(filename)
		nameWithoutExt := strings.TrimSuffix(filename, ext)
		outputDir := dir
		if strings.TrimSpace(lib.Config.CacheDir) != "" {
			outputDir = lib.Config.CacheDir
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				log.Printf("Failed to create cache directory: %s", err.Error())
				continue
			}
		}
		outputPath := filepath.Join(outputDir, nameWithoutExt+".tmp"+ext)

		if err := transcode(lib.Config, path, outputPath); err != nil {
			SaveHistory(db, logMsg(fmt.Sprintf("Failed to transcode: %s", err.Error())))
			os.Remove(outputPath)
			continue
		}

		if lib.Config.Filesize {
			outputInfo, err := os.Stat(outputPath)
			if err != nil {
				log.Printf("Failed to get output file info: %s", err.Error())
				continue
			}
			if outputInfo.Size() >= info.Size() {
				SaveHistory(db, logMsg(fmt.Sprintf("Transcoded file is not smaller, skipping replacement: %s", path)))
				os.Remove(outputPath)
				err = addSkip(db, id, path, "Transcoded file is not smaller")
				if err != nil {
					log.Printf("Failed to add to skiplist: %s", err.Error())
				}
				continue
			}
		}

		if err := os.Remove(path); err != nil {
			log.Printf("Failed to remove original file: %s", err.Error())
			continue
		}

		if err := os.Rename(outputPath, path); err != nil {
			log.Printf("Failed to rename transcoded file: %s", err.Error())
			continue
		}

		err = addSkip(db, id, path, "Successfully transcoded")
		if err != nil {
			log.Printf("Failed to add to skiplist: %s", err.Error())
		}
	}

}

func RunJob(db *sql.DB, js *JobScheduler, id int) {
	js.runJob(db, id)
}
