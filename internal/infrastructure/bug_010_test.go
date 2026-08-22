package infrastructure

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBug010_CopyFileKeepsWrittenDestination(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source.txt")
	destination := filepath.Join(root, "destination.txt")
	if err := os.WriteFile(source, []byte("payload"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := copyFile(source, destination); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("copied destination is unavailable: %v", err)
	}
	if string(got) != "payload" {
		t.Fatalf("copied destination changed: %q", got)
	}
}
