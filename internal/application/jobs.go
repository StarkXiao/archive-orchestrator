package application

import (
	"archive-orchestrator/internal/domain"
	"archive-orchestrator/internal/infrastructure"
	"context"
	"time"
)

type Jobs struct{ db *infrastructure.Store }

func NewJobs(db *infrastructure.Store) *Jobs { return &Jobs{db: db} }
func (j *Jobs) Create(c context.Context, r domain.Rule, t domain.JobType) (domain.Job, error) {
	v := domain.Job{ID: infrastructure.NewID("job"), RuleID: r.ID, Type: t, Status: domain.Pending, RuleSnapshot: r, ScheduledAt: time.Now(), CreatedAt: time.Now()}
	return v, j.db.SaveJob(c, v)
}
func (j *Jobs) Get(c context.Context, id string) (domain.Job, error) { return j.db.GetJob(c, id) }
func (j *Jobs) List(c context.Context) ([]domain.Job, error)         { return j.db.ListJobs(c) }
func (j *Jobs) Retry(c context.Context, id string) (domain.Job, error) {
	v, e := j.Get(c, id)
	if e != nil {
		return v, e
	}
	if v.Status != domain.Failed && v.Status != domain.ManualReview && v.Status != domain.PartialFailed {
		return v, domain.ErrConflict
	}
	v.Status = domain.RetryWait
	v.Error = ""
	return v, j.db.SaveJob(c, v)
}
func (j *Jobs) Cancel(c context.Context, id string) (domain.Job, error) {
	v, e := j.Get(c, id)
	if e != nil {
		return v, e
	}
	if v.Status == domain.Running || v.Terminal() {
		return v, domain.ErrConflict
	}
	v.Status = domain.Cancelled
	return v, j.db.SaveJob(c, v)
}
func (j *Jobs) Events(c context.Context, id string) ([]domain.Event, error) {
	return j.db.Events(c, id)
}

func (j *Jobs) Fail(c context.Context, v domain.Job, cause error) error {
	now := time.Now()
	v.Error = cause.Error()
	v.FinishedAt = &now
	if v.Attempts >= v.RuleSnapshot.MaxAttempts {
		v.Status = domain.ManualReview
		v.RetryAt = nil
		return j.db.SaveJob(c, v)
	}
	delay := time.Second * time.Duration(1<<min(v.Attempts, 6))
	retryAt := now.Add(delay)
	v.Status = domain.RetryWait
	v.RetryAt = &retryAt
	return j.db.SaveJob(c, v)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
