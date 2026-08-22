package application

import (
	"archive-orchestrator/internal/domain"
	"archive-orchestrator/internal/infrastructure"
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func TestBug004_CreatePreservesInvalidRuleError(t *testing.T) {
	db, err := infrastructure.NewStore(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewRules(db).Create(context.Background(), domain.Rule{})
	if !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("invalid rule error identity was lost: %v", err)
	}
}
