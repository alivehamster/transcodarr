package internal

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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
fileLoop:
	for _, path := range files {

		currentPath := path

		// Check if the file should be skipped
		if _, shouldSkip := skipMap[path]; shouldSkip {
			SaveHistory(db, logMsg(fmt.Sprintf("Skipping: %s", path)))
			continue
		}

		for _, filter := range lib.Config.Order {
			switch filter.ID {
			case "fileAge":
				if !FileAgeFilter(id, filter, currentPath, db) {
					continue fileLoop
				}
			case "minimumFileSize":
				if !MinSizeFilter(id, filter, currentPath, db) {
					continue fileLoop
				}
			case "hardlinks":
				if !HardlinkFilter(id, filter, currentPath, db) {
					continue fileLoop
				}
			case "mediaCodec":
				if !CodecFilter(id, filter, currentPath, db) {
					continue fileLoop
				}
			case "newFileSize":
				if !NewFileSizeFilter(id, filter, path, currentPath, db) {
					continue fileLoop
				}
			case "bitrate":
				if !BitrateFilter(id, filter, currentPath, db) {
					continue fileLoop
				}
			case "transcode":

				SaveHistory(db, logMsg(fmt.Sprintf("Processing: %s", path)))

				filename := filepath.Base(path)
				dir := filepath.Dir(path)
				ext := filepath.Ext(filename)
				nameWithoutExt := strings.TrimSuffix(filename, ext)
				outputDir := dir
				if strings.TrimSpace(lib.Config.CacheDir) != "" {
					outputDir = lib.Config.CacheDir
					if err := os.MkdirAll(outputDir, 0755); err != nil {
						log.Printf("Failed to create cache directory: %s", err.Error())
						continue fileLoop
					}
				}
				currentPath := filepath.Join(outputDir, nameWithoutExt+".tmp"+ext)

				if err := transcode(lib.Config, path, currentPath); err != nil {
					SaveHistory(db, logMsg(fmt.Sprintf("Failed to transcode: %s", err.Error())))
					os.Remove(currentPath)
					continue fileLoop
				}

			default:
				log.Printf("Unknown filter ID: %s", filter.ID)
			}
		}

		if err := replaceFile(currentPath, path); err != nil {
			log.Printf("Failed to replace original file: %s", err.Error())
			continue fileLoop
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
