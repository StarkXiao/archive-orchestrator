package worker

import (
	"archive-orchestrator/internal/domain"
	"testing"
	"time"
)

func TestBug002_RetryWaitWithoutTimestampIsClaimable(t *testing.T) {
	job := domain.Job{Status: domain.RetryWait}
	if !readyForClaim(job, time.Now()) {
		t.Fatal("retry job without a retry timestamp was not claimable")
	}
}
