package worker

import (
	"archive-orchestrator/internal/domain"
	"archive-orchestrator/internal/infrastructure"
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestBug007_UnexpiredOrMissingLeaseIsNotRequeued(t *testing.T) {
	db, err := infrastructure.NewStore(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	job := domain.Job{ID: "missing", Status: domain.Running}
	if err := db.SaveJob(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	if err := db.RequeueExpired(time.Now()); err != nil {
		t.Fatal(err)
	}
	got, err := db.GetJob(context.Background(), job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.Running {
		t.Fatalf("job without lease was requeued: %s", got.Status)
	}
}
