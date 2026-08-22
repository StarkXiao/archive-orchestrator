package application

import (
	"archive-orchestrator/internal/domain"
	"archive-orchestrator/internal/infrastructure"
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func newBug009TestStore(t *testing.T) *infrastructure.Store {
	t.Helper()
	dir := t.TempDir()
	s, err := infrastructure.NewStore(filepath.Join(dir, "state", "store.json"))
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	return s
}

// TestBug009_InvalidPatternRetainsInvalidError mirrors the acceptance
// regression: an unclosed archive wildcard must be rejected at creation, and
// the caller must identify the failure via errors.Is(err, domain.ErrInvalid).
func TestBug009_InvalidPatternRetainsInvalidError(t *testing.T) {
	store := newBug009TestStore(t)
	rules := NewRules(store)

	// Unclosed character class — a syntax error for filepath.Match.
	r := domain.Rule{
		Name:             "unclosed",
		SourceRoot:       "/data/src",
		ArchiveRoot:      "/archive/arc",
		Pattern:          "foo[bar",
		RetentionDays:    1,
		IntervalMinutes:  1,
		MaxFiles:         1,
		MaxAttempts:      1,
		Enabled:          true,
	}

	_, err := rules.Create(context.Background(), r)
	if err == nil {
		t.Fatalf("expected creation to be rejected for invalid pattern, got nil error")
	}
	if !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("expected errors.Is(err, domain.ErrInvalid)=true; got err=%v", err)
	}
	// Context about the offending pattern should be preserved.
	if want := "foo[bar"; !strings.Contains(err.Error(), want) {
		t.Fatalf("expected error to retain pattern context %q; got %q", want, err.Error())
	}
}

// TestBug009_LegalPatternAccepted ensures legal wildcards and valid rule
// creation behavior are unaffected.
func TestBug009_LegalPatternAccepted(t *testing.T) {
	store := newBug009TestStore(t)
	rules := NewRules(store)

	r := domain.Rule{
		Name:             "legal",
		SourceRoot:       "/data/src",
		ArchiveRoot:      "/archive/arc",
		Pattern:          "*.log",
		RetentionDays:    1,
		IntervalMinutes:  1,
		MaxFiles:         1,
		MaxAttempts:      1,
		Enabled:          true,
	}

	created, err := rules.Create(context.Background(), r)
	if err != nil {
		t.Fatalf("expected legal rule to be created; got err=%v", err)
	}
	if created.ID == "" {
		t.Fatalf("expected created rule to receive an ID")
	}
}

// TestBug009_OtherValidationUnaffected checks the other rule validation paths
// still reject and remain identifiable as domain.ErrInvalid.
func TestBug009_OtherValidationUnaffected(t *testing.T) {
	store := newBug009TestStore(t)
	rules := NewRules(store)

	// Missing required field (empty pattern).
	r := domain.Rule{
		Name:            "empty-pattern",
		SourceRoot:      "/data/src",
		ArchiveRoot:     "/archive/arc",
		Pattern:         "",
		IntervalMinutes: 1,
		MaxFiles:        1,
		MaxAttempts:     1,
	}
	if _, err := rules.Create(context.Background(), r); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("expected ErrInvalid for empty pattern; got err=%v", err)
	}
}
