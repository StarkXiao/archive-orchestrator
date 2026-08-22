package application

import (
	"archive-orchestrator/internal/domain"
	"archive-orchestrator/internal/infrastructure"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestBug005_CancelledArchiveDoesNotCreateBatch(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "old.log"), []byte("data"), 0600); err != nil {
		t.Fatal(err)
	}
	db, err := infrastructure.NewStore(filepath.Join(root, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	job := domain.Job{ID: "job-1", RuleSnapshot: domain.Rule{SourceRoot: root, ArchiveRoot: filepath.Join(root, "archive"), Pattern: "*.log", RetentionDays: 0}}
	_, err = NewArchive(db, infrastructure.NewFiles(db), NewJobs(db)).Run(ctx, job)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled archive continued: %v", err)
	}
}
