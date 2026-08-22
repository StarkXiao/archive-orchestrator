package application

import (
	"archive-orchestrator/internal/domain"
	"archive-orchestrator/internal/infrastructure"
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func TestBug006_RestoreRejectsSiblingOfConfiguredRoot(t *testing.T) {
	root := t.TempDir()
	db, err := infrastructure.NewStore(filepath.Join(root, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	batch := domain.Batch{ID: "batch-1", Root: filepath.Join(root, "archive")}
	if err := db.SaveBatch(context.Background(), batch); err != nil {
		t.Fatal(err)
	}
	batches := NewBatches(db, infrastructure.NewFiles(db), NewJobs(db), filepath.Join(root, "restore"))
	_, err = batches.Restore(context.Background(), batch.ID, filepath.Join(root, "restore-other"), "skip")
	if !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("sibling restore target was accepted: %v", err)
	}
}
