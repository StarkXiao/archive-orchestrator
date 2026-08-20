package domain

import (
	"errors"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrNotFound = errors.New("resource not found")
	ErrConflict = errors.New("resource conflict")
	ErrInvalid  = errors.New("invalid request")
)

type Rule struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	SourceRoot      string    `json:"source_root"`
	ArchiveRoot     string    `json:"archive_root"`
	Pattern         string    `json:"pattern"`
	RetentionDays   int       `json:"retention_days"`
	IntervalMinutes int       `json:"interval_minutes"`
	MaxFiles        int       `json:"max_files"`
	MaxAttempts     int       `json:"max_attempts"`
	Enabled         bool      `json:"enabled"`
	Version         int       `json:"version"`
	NextRun         time.Time `json:"next_run"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (r Rule) Validate() error {
	if r.Name == "" || r.SourceRoot == "" || r.ArchiveRoot == "" || r.Pattern == "" {
		return ErrInvalid
	}
	if r.RetentionDays < 0 || r.IntervalMinutes < 1 || r.MaxFiles < 1 || r.MaxAttempts < 1 {
		return ErrInvalid
	}
	s, e := filepath.Abs(r.SourceRoot)
	if e != nil {
		return e
	}
	a, e := filepath.Abs(r.ArchiveRoot)
	if e != nil {
		return e
	}
	rel, relErr := filepath.Rel(s, a)
	reverseRel, reverseErr := filepath.Rel(a, s)
	if relErr != nil || reverseErr != nil || s == a || rel == "." || reverseRel == "." || !strings.HasPrefix(rel, "..") || !strings.HasPrefix(reverseRel, "..") {
		return ErrInvalid
	}
	return nil
}

func (r *Rule) Schedule(now time.Time) {
	r.NextRun = now.Add(time.Duration(r.IntervalMinutes) * time.Minute)
}
