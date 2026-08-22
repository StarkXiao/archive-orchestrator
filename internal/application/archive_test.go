package application

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"archive-orchestrator/internal/domain"
	"archive-orchestrator/internal/infrastructure"
)

// TestArchiveRunPreservesErrorType asserts that a failure originating in the
// infrastructure layer is wrapped (not flattened) by Archive.Run so callers
// can still unwrap to the original concrete error type.
func TestArchiveRunPreservesErrorType(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	if err := os.MkdirAll(source, 0755); err != nil {
		t.Fatal(err)
	}
	// An old file so the retention walk picks it up and Run reaches Create.
	src := filepath.Join(source, "old.txt")
	if err := os.WriteFile(src, []byte("data"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(src, time.Unix(0, 0), time.Unix(0, 0)); err != nil {
		t.Fatal(err)
	}
	// Point ArchiveRoot at a regular file so MkdirAll fails with ENOTDIR,
	// producing a concrete *fs.PathError that must survive the wrap.
	archiveRoot := filepath.Join(root, "not-a-dir")
	if err := os.WriteFile(archiveRoot, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}

	store, err := infrastructure.NewStore(filepath.Join(root, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	files := infrastructure.NewFiles(store)
	jobs := NewJobs(store)
	archive := NewArchive(store, files, jobs)

	rule := domain.Rule{
		ID:            "rule-1",
		Name:          "test",
		SourceRoot:    source,
		ArchiveRoot:   archiveRoot,
		Pattern:       "*.txt",
		RetentionDays: 1,
		MaxFiles:      10,
		MaxAttempts:   3,
	}
	job := domain.Job{ID: "job-1", Type: domain.JobArchive, RuleSnapshot: rule, ScheduledAt: time.Now()}

	_, err = archive.Run(context.Background(), job)
	if err == nil {
		t.Fatal("archive.Run: want error, got nil")
	}
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("error type not preserved through wrap: %T: %v", err, err)
	}
}
