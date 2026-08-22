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

// TestRestoreRejectsTamperedArchiveFile ensures the integrity gate runs before the
// restore rename: if an archived entry's bytes/digest no longer match the manifest,
// the temp file must be dropped and counted as failed, never committed to the target.
func TestRestoreRejectsTamperedArchiveFile(t *testing.T) {
	root := t.TempDir()
	source, archive := filepath.Join(root, "source"), filepath.Join(root, "archive")
	src := filepath.Join(source, "data.txt")
	if err := os.MkdirAll(filepath.Dir(src), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(src, []byte("original-content"), 0600); err != nil {
		t.Fatal(err)
	}
	store, err := NewStore(filepath.Join(root, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	files := NewFiles(store)
	batch, err := files.Create(context.Background(), domain.Rule{SourceRoot: source, ArchiveRoot: archive}, "job-1",
		[]domain.ManifestEntry{{Source: src, ModifiedAt: time.Now()}})
	if err != nil {
		t.Fatal(err)
	}

	// Tamper with the archived file after creation so its size and SHA256 no
	// longer match the manifest, simulating a replaced/corrupted archive entry.
	if err := os.WriteFile(batch.Entries[0].Archive, []byte("tampered"), 0600); err != nil {
		t.Fatal(err)
	}

	restoreRoot := filepath.Join(root, "restore")
	report, err := files.Restore(context.Background(), batch, restoreRoot, "skip")
	if err != nil {
		t.Fatal(err)
	}
	if report.Failed != 1 || report.Restored != 0 {
		t.Fatalf("tampered entry must be rejected, got report: %+v", report)
	}
	// The committed target must not exist, and the leftover temp file must be removed.
	dst := filepath.Join(restoreRoot, "data.txt")
	if _, err := os.Stat(dst); err == nil {
		t.Fatalf("tampered file was committed to target: %s", dst)
	}
	if _, err := os.Stat(dst + ".restore-tmp"); err == nil {
		t.Fatalf("temp file was not cleaned up: %s", dst+".restore-tmp")
	}
}
