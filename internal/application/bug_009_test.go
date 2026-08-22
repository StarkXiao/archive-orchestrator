package application

import (
	"archive-orchestrator/internal/domain"
	"archive-orchestrator/internal/infrastructure"
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func TestBug009_InvalidPatternErrorSurvivesApplicationBoundary(t *testing.T) {
	root := t.TempDir()
	db, err := infrastructure.NewStore(filepath.Join(root, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	rule := domain.Rule{Name: "daily", SourceRoot: filepath.Join(root, "source"), ArchiveRoot: filepath.Join(root, "archive"), Pattern: "[", IntervalMinutes: 1, MaxFiles: 1, MaxAttempts: 1}
	_, err = NewRules(db).Create(context.Background(), rule)
	if !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("invalid pattern lost its error identity: %v", err)
	}
}
