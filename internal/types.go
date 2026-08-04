package internal

import (
	"encoding/json"
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
	Dirs              []string        `json:"dirs"`
	HandbrakeCategory string          `json:"handbrakeCategory"`
	HandbrakeProfile  string          `json:"handbrakeProfile"`
	CacheDir          string          `json:"cacheDir"`
	Nodes             json.RawMessage `json:"nodes"`
	Edges             json.RawMessage `json:"edges"`
	Order             []Order         `json:"order"`
}
type Order struct {
	ID         string   `json:"id"`
	Int        int      `json:"data_int"`
	SkipFuture bool     `json:"skipFuture"`
	Array      []string `json:"data_array"`
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
