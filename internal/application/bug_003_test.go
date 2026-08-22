package application

import (
	"archive-orchestrator/internal/domain"
	"archive-orchestrator/internal/infrastructure"
	"context"
	"path/filepath"
	"testing"
)

func TestBug003_ListRulesDoesNotMutateStoredNames(t *testing.T) {
	db, err := infrastructure.NewStore(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	want := domain.Rule{ID: "rule-1", Name: "  daily archive  "}
	if err := db.SaveRule(context.Background(), want); err != nil {
		t.Fatal(err)
	}
	rules := NewRules(db)
	if _, err := rules.List(context.Background()); err != nil {
		t.Fatal(err)
	}
	got, err := rules.Get(context.Background(), want.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != want.Name {
		t.Fatalf("list mutated stored name: got %q want %q", got.Name, want.Name)
	}
}
