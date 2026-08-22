package infrastructure

import (
	"archive-orchestrator/internal/domain"
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestBug001_RenewLeaseRejectsForeignWorker(t *testing.T) {
	store, err := NewStore(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	job := domain.Job{ID: "job-1", Status: domain.Pending}
	if err := store.SaveJob(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Claim(context.Background(), job.ID, "worker-a"); err != nil {
		t.Fatal(err)
	}
	if err := store.RenewLease(job.ID, "worker-b", time.Now()); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("foreign worker renewed lease: %v", err)
	}
}
