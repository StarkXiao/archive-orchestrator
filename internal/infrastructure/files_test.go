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

func TestCreateRejectsSourcesOutsideRoot(t *testing.T) {
	root := t.TempDir()
	source, archive := filepath.Join(root, "source"), filepath.Join(root, "archive")
	sibling := filepath.Join(root, "source-other")
	for _, p := range []string{
		filepath.Join(sibling, "report.txt"), // sibling of source root
		filepath.Join(sibling, "x", "report.txt"),
		filepath.Join(root, "report.txt"), // parent of source root
	} {
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("data"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	store, err := NewStore(filepath.Join(root, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	files := NewFiles(store)
	cases := []string{
		filepath.Join(sibling, "report.txt"),
		filepath.Join(sibling, "x", "report.txt"),
		filepath.Join(root, "report.txt"),
	}
	for i, src := range cases {
		_, err := files.Create(context.Background(),
			domain.Rule{SourceRoot: source, ArchiveRoot: archive}, "job-1",
			[]domain.ManifestEntry{{Source: src, ModifiedAt: time.Now()}})
		if err == nil {
			t.Fatalf("case %d src=%s: expected error for out-of-root source", i, src)
		}
	}
}

func TestCreateAllowsNestedSourcesInsideRoot(t *testing.T) {
	root := t.TempDir()
	source, archive := filepath.Join(root, "source"), filepath.Join(root, "archive")
	src := filepath.Join(source, "deep", "nested", "report.txt")
	if err := os.MkdirAll(filepath.Dir(src), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(src, []byte("data"), 0600); err != nil {
		t.Fatal(err)
	}
	store, err := NewStore(filepath.Join(root, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	files := NewFiles(store)
	batch, err := files.Create(context.Background(),
		domain.Rule{SourceRoot: source, ArchiveRoot: archive}, "job-1",
		[]domain.ManifestEntry{{Source: src, ModifiedAt: time.Now()}})
	if err != nil {
		t.Fatalf("nested in-root source should be allowed, got %v", err)
	}
	if got := batch.Entries[0].Archive; !filepath.IsAbs(got) {
		t.Fatalf("archive path not absolute: %s", got)
	}
}
