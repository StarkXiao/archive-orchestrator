package application

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"archive-orchestrator/internal/domain"
	"archive-orchestrator/internal/infrastructure"
)

func newBatchesStore(t *testing.T) (*infrastructure.Store, string) {
	t.Helper()
	root := t.TempDir()
	store, err := infrastructure.NewStore(filepath.Join(root, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	return store, root
}

func TestRestoreRejectsSiblingAndParentTargets(t *testing.T) {
	store, root := newBatchesStore(t)
	restoreRoot := filepath.Join(root, "restore")
	if err := os.MkdirAll(restoreRoot, 0755); err != nil {
		t.Fatal(err)
	}
	// seed a batch to restore
	files := infrastructure.NewFiles(store)
	archiveRoot := filepath.Join(root, "archive")
	entries := []domain.ManifestEntry{{
		Source: filepath.Join(root, "src", "report.txt"),
	}}
	if err := os.MkdirAll(filepath.Dir(entries[0].Source), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(entries[0].Source, []byte("data"), 0600); err != nil {
		t.Fatal(err)
	}
	batch, err := files.Create(context.Background(),
		domain.Rule{SourceRoot: filepath.Join(root, "src"), ArchiveRoot: archiveRoot},
		"job-1", entries)
	if err != nil {
		t.Fatal(err)
	}
	jobs := NewJobs(store)
	b := NewBatches(store, files, jobs, restoreRoot)

	cases := []string{
		filepath.Join(root, "restore-other"), // sibling of restore root
		filepath.Join(root, "restore-other", "x"),
		root, // parent of restore root
	}
	for i, target := range cases {
		if _, err := b.Restore(context.Background(), batch.ID, target, "skip"); err != domain.ErrInvalid {
			t.Fatalf("case %d target=%s: expected ErrInvalid, got %v", i, target, err)
		}
	}
}

func TestRestoreAllowsNestedSubdirInsideRoot(t *testing.T) {
	store, root := newBatchesStore(t)
	restoreRoot := filepath.Join(root, "restore")
	if err := os.MkdirAll(restoreRoot, 0755); err != nil {
		t.Fatal(err)
	}
	files := infrastructure.NewFiles(store)
	archiveRoot := filepath.Join(root, "archive")
	entries := []domain.ManifestEntry{{
		Source: filepath.Join(root, "src", "a", "b", "report.txt"),
	}}
	if err := os.MkdirAll(filepath.Dir(entries[0].Source), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(entries[0].Source, []byte("data"), 0600); err != nil {
		t.Fatal(err)
	}
	batch, err := files.Create(context.Background(),
		domain.Rule{SourceRoot: filepath.Join(root, "src"), ArchiveRoot: archiveRoot},
		"job-1", entries)
	if err != nil {
		t.Fatal(err)
	}
	jobs := NewJobs(store)
	b := NewBatches(store, files, jobs, restoreRoot)

	target := filepath.Join(restoreRoot, "deep", "nested", "dir")
	if _, err := b.Restore(context.Background(), batch.ID, target, "skip"); err != nil {
		t.Fatalf("nested subdir restore should be allowed, got %v", err)
	}
}
