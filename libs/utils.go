package libs

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var videoExtensions = map[string]struct{}{
	".mkv":  {},
	".mp4":  {},
	".m4v":  {},
	".avi":  {},
	".mov":  {},
	".wmv":  {},
	".mpg":  {},
	".mpeg": {},
	".ts":   {},
	".m2ts": {},
	".mts":  {},
	".vob":  {},
	".flv":  {},
	".webm": {},
}

func getlibItems(lib Library) []string {
	var paths []string

	for _, dir := range lib.Config.Dirs {
		filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if _, ok := videoExtensions[ext]; ok && !strings.Contains(filepath.Base(path), ".tmp.") {
				paths = append(paths, path)
			}
			return nil
		})
	}

	return paths
}

func getCodec(path string) (string, error) {
	out, err := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=codec_name",
		"-of", "json",
		path,
	).Output()
	if err != nil {
		return "", err
	}
	var result struct {
		Streams []struct {
			CodecName string `json:"codec_name"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return "", err
	}
	if len(result.Streams) == 0 {
		return "", nil
	}
	return result.Streams[0].CodecName, nil
}

func logMsg(msg string) string {
	log.Println(msg)
	return fmt.Sprintf("%s %s", time.Now().Format(time.DateTime), msg)
}

func updateSkiplist(db *sql.DB, id int, skiplist []Skip, entry Skip) ([]Skip, error) {
	skiplist = append(skiplist, entry)

	updated, err := json.Marshal(skiplist)
	if err != nil {
		log.Printf("Failed to serialize skiplist: %s", err.Error())
		return skiplist, err
	}
	if _, err := db.Exec("UPDATE libraries SET skiplist = ? WHERE id = ?", string(updated), id); err != nil {
		log.Printf("Failed to update skiplist: %s", err.Error())
		return skiplist, err
	}
	return skiplist, nil
}

func SaveHistory(db *sql.DB, text string) {
	if _, err := db.Exec("INSERT INTO history (text) VALUES (?)", text); err != nil {
		log.Printf("Failed to save history: %s", err.Error())
	}
}
