package worker

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"archive-orchestrator/internal/application"
	"archive-orchestrator/internal/domain"
	"archive-orchestrator/internal/infrastructure"
)

// TestRestoreExecutorCountsFailuresInJobFiles ensures the executor propagates the
// full set of processed files into the job state. A restore that fails the
// per-file integrity gate must record the failed entry in j.Files (not just in
// the error message), and land in PartialFailed.
func TestRestoreExecutorCountsFailuresInJobFiles(t *testing.T) {
	root := t.TempDir()
	source, archive := filepath.Join(root, "source"), filepath.Join(root, "archive")
	src := filepath.Join(source, "data.txt")
	if err := os.MkdirAll(filepath.Dir(src), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(src, []byte("original-content"), 0600); err != nil {
		t.Fatal(err)
	}
	store, err := infrastructure.NewStore(filepath.Join(root, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	files := infrastructure.NewFiles(store)
	jobs := application.NewJobs(store)
	batches := application.NewBatches(store, files, jobs, root)
	archiveApp := application.NewArchive(store, files, jobs)
	exec := NewExecutor(jobs, archiveApp, batches, store)

	batch, err := files.Create(context.Background(), domain.Rule{SourceRoot: source, ArchiveRoot: archive}, "job-create",
		[]domain.ManifestEntry{{Source: src, ModifiedAt: time.Now()}})
	if err != nil {
		t.Fatal(err)
	}
	// Tamper so the restore integrity gate rejects the single entry.
	if err := os.WriteFile(batch.Entries[0].Archive, []byte("tampered"), 0600); err != nil {
		t.Fatal(err)
	}

	restoreRoot := filepath.Join(root, "restore")
	job := domain.Job{
		ID:        "job-restore",
		RuleID:    batch.ID,
		Type:      domain.JobRestore,
		Status:    domain.Running,
		RuleSnapshot: domain.Rule{ID: batch.ID, ArchiveRoot: batch.Root, MaxAttempts: 3},
		RestoreTarget: restoreRoot,
		ConflictMode:  "skip",
	}
	if err := store.SaveJob(context.Background(), job); err != nil {
		t.Fatal(err)
	}

	if err := exec.Run(context.Background(), job); err != nil {
		t.Fatalf("exec.Run: %v", err)
	}
	saved, err := store.GetJob(context.Background(), job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Status != domain.PartialFailed {
		t.Fatalf("status = %s, want %s", saved.Status, domain.PartialFailed)
	}
	// One file was processed and it failed; the count must reflect it.
	if saved.Files != 1 {
		t.Fatalf("Files = %d, want 1 (failed entries must be counted, not dropped)", saved.Files)
	}
	// And the target must not have been committed.
	if _, err := os.Stat(filepath.Join(restoreRoot, "data.txt")); err == nil {
		t.Fatal("tampered file was committed to restore target")
	}
}
