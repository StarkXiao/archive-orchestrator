package infrastructure

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"archive-orchestrator/internal/domain"
)

func TestCreatePreservesRelativePathsAndRestores(t *testing.T) {
	root := t.TempDir()
	source, archive := filepath.Join(root, "source"), filepath.Join(root, "archive")
	for _, name := range []string{"a/report.txt", "b/report.txt"} {
		path := filepath.Join(source, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(name), 0600); err != nil {
			t.Fatal(err)
		}
	}
	store, err := NewStore(filepath.Join(root, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	files := NewFiles(store)
	entries := []domain.ManifestEntry{{Source: filepath.Join(source, "a/report.txt"), ModifiedAt: time.Now()}, {Source: filepath.Join(source, "b/report.txt"), ModifiedAt: time.Now()}}
	batch, err := files.Create(context.Background(), domain.Rule{SourceRoot: source, ArchiveRoot: archive}, "job-1", entries)
	if err != nil {
		t.Fatal(err)
	}
	if batch.Entries[0].Archive == batch.Entries[1].Archive {
		t.Fatal("same-named files were overwritten")
	}
	report, err := files.Verify(context.Background(), batch)
	if err != nil || report.Bad != 0 {
		t.Fatalf("verify: %+v, %v", report, err)
	}
	restoreRoot := filepath.Join(root, "restore")
	restored, err := files.Restore(context.Background(), batch, restoreRoot, "skip")
	if err != nil || restored.Restored != 2 {
		t.Fatalf("restore: %+v, %v", restored, err)
	}
	for _, name := range []string{"a/report.txt", "b/report.txt"} {
		if _, err := os.Stat(filepath.Join(restoreRoot, name)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCopyFileKeepsDestinationOnSuccess(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "in.txt")
	dst := filepath.Join(dir, "out.txt")
	if err := os.WriteFile(src, []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile: %v", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("destination removed after successful copy: %v", err)
	}
	if string(got) != "hello" {
		t.Fatalf("destination content = %q, want %q", got, "hello")
	}
}

func TestCopyFileRemovesDestinationOnFailure(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "missing.txt")
	dst := filepath.Join(dir, "out.txt")
	if err := copyFile(src, dst); err == nil {
		t.Fatal("copyFile with missing source: want error, got nil")
	}
	if _, err := os.Stat(dst); err == nil {
		t.Fatal("destination left behind after failed copy")
	}
}

