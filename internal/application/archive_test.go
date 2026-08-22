package application

import (
	"archive-orchestrator/internal/domain"
	"archive-orchestrator/internal/infrastructure"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestArchiveRunAbortsOnCancelledContext guards against the caller's context
// being swapped for context.Background() between the application layer and the
// file store. When a request is cancelled before it reaches the service, the
// cancellation must propagate so no temp batch is created, no files are copied
// and no batch state is persisted.
func TestArchiveRunAbortsOnCancelledContext(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	archive := filepath.Join(root, "archive")
	if err := os.MkdirAll(filepath.Join(source, "a"), 0755); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(source, "a", "report.txt")
	if err := os.WriteFile(src, []byte("payload"), 0600); err != nil {
		t.Fatal(err)
	}

	store, err := infrastructure.NewStore(filepath.Join(root, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	files := infrastructure.NewFiles(store)
	jobs := NewJobs(store)
	archiveSvc := NewArchive(store, files, jobs)

	rule := domain.Rule{
		ID:            "rule-1",
		SourceRoot:    source,
		ArchiveRoot:   archive,
		Pattern:       "report.txt",
		RetentionDays: 0,
		MaxFiles:      1,
		MaxAttempts:   3,
	}
	job := domain.Job{
		ID:           "job-1",
		RuleID:       rule.ID,
		Type:         domain.JobArchive,
		Status:       domain.Pending,
		RuleSnapshot: rule,
		CreatedAt:    time.Now(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := archiveSvc.Run(ctx, job); err == nil {
		t.Fatal("expected Run to abort with a cancelled context, got nil error")
	}

	// No temp batch root should have been materialised.
	matches, _ := filepath.Glob(archive + "*")
	for _, m := range matches {
		t.Errorf("unexpected batch artifact created despite cancellation: %s", m)
	}

	// No batch state should have been persisted.
	if batches, _ := store.ListBatches(context.Background()); len(batches) != 0 {
		t.Errorf("expected zero persisted batches, got %d", len(batches))
	}
}
