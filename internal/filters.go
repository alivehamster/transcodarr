package internal

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"slices"
	"syscall"
	"time"
)

func MinSizeFilter(id int, data Order, path string, db *sql.DB) bool {

	info, err := os.Stat(path)
	if err != nil {
		log.Printf("Error: %v", err)
		return false
	}

	minimumBytes := int64(data.Int) * 1024 * 1024
	if info.Size() < minimumBytes {
		SaveHistory(db, logMsg(fmt.Sprintf("Skipping file smaller than minimum size (%d MB): %s", data.Int, path)))
		if data.SkipFuture {
			err = addSkip(db, id, path, fmt.Sprintf("File is smaller than minimum size (%d MB)", data.Int))
			if err != nil {
				log.Printf("Failed to add to skiplist: %s", err.Error())
			}
		}
		return false
	}
	return true
}

func HardlinkFilter(id int, data Order, path string, db *sql.DB) bool {

	info, err := os.Stat(path)
	if err != nil {
		log.Printf("Error: %v", err)
		return false
	}

	// Cast the Sys() interface to the platform-specific Stat_t type
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		log.Println("Not a Unix-like system; cannot check hardlinks")
		return false
	}

	if stat.Nlink > 1 {
		SaveHistory(db, logMsg(fmt.Sprintf("Skipping file with multiple hardlinks: %s", path)))
		if data.SkipFuture {
			err = addSkip(db, id, path, "File has multiple hardlinks")
			if err != nil {
				log.Printf("Failed to add to skiplist: %s", err.Error())
			}
		}
		return false
	}

	return true
}

func FileAgeFilter(id int, data Order, path string, db *sql.DB) bool {

	info, err := os.Stat(path)
	if err != nil {
		log.Printf("Error: %v", err)
		return false
	}
	if time.Since(info.ModTime()) < (time.Duration(data.Int) * 24 * time.Hour) {
		SaveHistory(db, logMsg(fmt.Sprintf("Skipping recently changed file: %s", path)))
		if data.SkipFuture {
			err = addSkip(db, id, path, fmt.Sprintf("File is newer than %d days", data.Int))
			if err != nil {
				log.Printf("Failed to add to skiplist: %s", err.Error())
			}
		}
		return false
	}
	return true
}

func CodecFilter(id int, data Order, path string, db *sql.DB) bool {

	codec, err := getCodec(path)
	if err != nil {
		log.Printf("Failed to get codec for %s: %s", path, err.Error())
		return false
	}
	if slices.Contains(data.Array, codec) {
		SaveHistory(db, logMsg(fmt.Sprintf("Skipping file with codec %s: %s", codec, path)))

		if data.SkipFuture {
			err = addSkip(db, id, path, fmt.Sprintf("Codec is already %s", codec))
			if err != nil {
				log.Printf("Failed to add to skiplist: %s", err.Error())
			}
		}
		return false
	}

	return true
}

func NewFileSizeFilter(id int, data Order, path string, outputPath string, db *sql.DB) bool {

	info, err := os.Stat(path)
	if err != nil {
		log.Printf("Error: %v", err)
		return false
	}

	outputInfo, err := os.Stat(outputPath)
	if err != nil {
		log.Printf("Failed to get output file info: %s", err.Error())
		return false
	}
	if outputInfo.Size() >= info.Size() {
		SaveHistory(db, logMsg(fmt.Sprintf("Transcoded file is not smaller, skipping replacement: %s", path)))
		os.Remove(outputPath)
		if data.SkipFuture {
			err = addSkip(db, id, path, "Transcoded file is not smaller")
			if err != nil {
				log.Printf("Failed to add to skiplist: %s", err.Error())
			}
		}
		return false
	}
	return true
}
