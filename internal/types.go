package internal

import (
	"sync"

	"github.com/robfig/cron/v3"
)

type Library struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Cron    string `json:"cron"`
	Config  Config `json:"config"`
	NextRun string `json:"nextRun"`
}

type Config struct {
	Dirs              []string `json:"dirs"`
	HandbrakeCategory string   `json:"handbrakeCategory"`
	HandbrakeProfile  string   `json:"handbrakeProfile"`
	CacheDir          string   `json:"cacheDir"`
	Order             []Order  `json:"order"`
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
	scheduler    *cron.Cron
	jobMap       map[int]cron.EntryID
	mu           sync.Mutex
	jobMu        sync.Mutex
	currentJobMu sync.RWMutex
	currentJob   int
}

type mediaInfoOutput struct {
	Media struct {
		Track []struct {
			Type           string `json:"@type"`
			Format         string `json:"Format"`
			CodecID        string `json:"CodecID"`
			OverallBitRate string `json:"OverallBitRate"` // bits/sec (on "General" track)
			BitRate        string `json:"BitRate"`        // bits/sec (on "Video" track)
		} `json:"track"`
	} `json:"media"`
}
