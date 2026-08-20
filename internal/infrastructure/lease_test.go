package infrastructure

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"archive-orchestrator/internal/domain"
)

func TestRenewLeaseKeepsRunningJobClaimed(t *testing.T) {
	store, err := NewStore(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	job := domain.Job{ID: "job-1", Status: domain.Pending}
	if err := store.SaveJob(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	claimed, err := store.Claim(context.Background(), job.ID, "worker-a")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.RenewLease(job.ID, "worker-a", time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := store.RequeueExpired(time.Now().Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	current, err := store.GetJob(context.Background(), job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if current.Status != domain.Running || current.LeaseOwner != claimed.LeaseOwner {
		t.Fatalf("job lease was lost: %+v", current)
	}
}
