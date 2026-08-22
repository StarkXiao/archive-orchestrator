package infrastructure

import (
	"archive-orchestrator/internal/domain"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestBug008_RestoreDoesNotCommitChecksumMismatch(t *testing.T) {
	root := t.TempDir()
	archive := filepath.Join(root, "entry.bin")
	if err := os.WriteFile(archive, []byte("corrupt"), 0600); err != nil {
		t.Fatal(err)
	}
	batch := domain.Batch{Root: root, Entries: []domain.ManifestEntry{{Source: "entry.bin", Archive: archive, Size: 4, SHA256: "wrong"}}}
	target := filepath.Join(root, "restore")
	report, err := NewFiles(nil).Restore(context.Background(), batch, target, "skip")
	if err != nil {
		t.Fatal(err)
	}
	if report.Failed != 1 || report.Restored != 0 {
		t.Fatalf("checksum mismatch was accepted: %+v", report)
	}
	if _, err := os.Stat(filepath.Join(target, "entry.bin")); !os.IsNotExist(err) {
		t.Fatalf("corrupt file was committed: %v", err)
	}
}
