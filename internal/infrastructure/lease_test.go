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
	// Reclaim comparison must use the real current time: a lease with ~2 minutes
	// left must survive a sweep taken at "now".
	if err := store.RequeueExpired(time.Now()); err != nil {
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

// A running job whose lease has not been written yet (LeaseUntil nil) must not be
// yanked back to retry_wait on the next sweep — that would re-execute a freshly
// claimed archive.
func TestRequeueExpiredSkipsRunningJobWithoutLease(t *testing.T) {
	store, err := NewStore(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	job := domain.Job{ID: "job-2", Status: domain.Running, LeaseOwner: "worker-a"}
	if err := store.SaveJob(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	if err := store.RequeueExpired(now); err != nil {
		t.Fatal(err)
	}
	current, err := store.GetJob(context.Background(), job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if current.Status != domain.Running {
		t.Fatalf("lease-missing job was requeued: %+v", current)
	}
}

// A genuinely expired lease is still requeued to retry_wait and becomes eligible
// to run again immediately (RetryAt == now). This preserves the retry behavior
// for explicitly expired tasks.
func TestRequeueExpiredRequeuesGenuinelyExpiredLease(t *testing.T) {
	store, err := NewStore(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	expired := now.Add(-time.Minute)
	job := domain.Job{ID: "job-3", Status: domain.Running, LeaseOwner: "worker-a", LeaseUntil: &expired}
	if err := store.SaveJob(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	if err := store.RequeueExpired(now); err != nil {
		t.Fatal(err)
	}
	current, err := store.GetJob(context.Background(), job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if current.Status != domain.RetryWait {
		t.Fatalf("expired lease was not requeued: %+v", current)
	}
	if current.RetryAt == nil || !current.RetryAt.Equal(now) {
		t.Fatalf("retry-at should be now for immediate retry: %+v", current.RetryAt)
	}
}
