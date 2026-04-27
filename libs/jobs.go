package libs

import (
	"database/sql"

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
	var name string
	err := db.QueryRow("SELECT name FROM libraries WHERE id = ?", id).Scan(&name)
	if err != nil {
		println("Failed to get library:", err.Error())
		return
	}
	println(name)
}
