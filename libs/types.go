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
	CacheDir          string   `json:"cacheDir"`
	FileAge           int      `json:"fileAge"`
	MinimumFileSizeMb int64    `json:"minimumFileSizeMb"`
	Hardlinks         bool     `json:"hardlinks"`
	MediaCodec        []string `json:"mediaCodec"`
	Filesize          bool     `json:"filesize"`
}

type Skip struct {
	ID          int    `json:"id"`
	Path        string `json:"path"`
	Description string `json:"description"`
}

type JobScheduler struct {
	scheduler *cron.Cron
	jobMap    map[int]cron.EntryID
	mu        sync.Mutex
	jobMu     sync.Mutex
}
