package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Listen      string
	DataFile    string
	Workers     int
	Poll        time.Duration
	Shutdown    time.Duration
	RestoreRoot string
}

func Load() Config {
	c := Config{Listen: ":8080", DataFile: "./data/state.json", RestoreRoot: "./data/restores", Workers: 2, Poll: 5 * time.Second, Shutdown: 10 * time.Second}
	if v := os.Getenv("ARCHIVE_ADDR"); v != "" {
		c.Listen = v
	}
	if v := os.Getenv("ARCHIVE_DATA"); v != "" {
		c.DataFile = v
	}
	if v := os.Getenv("ARCHIVE_RESTORE_ROOT"); v != "" {
		c.RestoreRoot = v
	}
	if v := os.Getenv("ARCHIVE_WORKERS"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 {
			c.Workers = n
		}
	}
	return c
}
