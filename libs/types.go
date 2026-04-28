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
	Dirs              []string `json:"dirs"`
	HandbrakeCategory string   `json:"handbrakeCategory"`
	HandbrakeProfile  string   `json:"handbrakeProfile"`
}

type Skip struct {
	Path        string `json:"path"`
	Description string `json:"description"`
}

type SkipList struct {
	ID    int    `json:"id"`
	Skips []Skip `json:"skips"`
}

type JobScheduler struct {
	scheduler *cron.Cron
	jobMap    map[int]cron.EntryID
	mu        sync.Mutex
}
