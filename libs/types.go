package libs

import (
	"sync"

	"github.com/robfig/cron/v3"
)

type Library struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Cron   string `json:"cron"`
	Config Config `json:"config"`
}

type Config struct {
	Dirs    []string `json:"dirs"`
	Profile string   `json:"profile"`
}

type JobScheduler struct {
	scheduler *cron.Cron
	jobMap    map[int]cron.EntryID
	mu        sync.Mutex
}
