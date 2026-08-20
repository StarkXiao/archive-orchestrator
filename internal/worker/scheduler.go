package worker

import (
	"archive-orchestrator/internal/application"
	"archive-orchestrator/internal/domain"
	"archive-orchestrator/internal/infrastructure"
	"context"
	"time"
)

type Scheduler struct {
	rules *application.Rules
	jobs  *application.Jobs
	db    *infrastructure.Store
}

func NewScheduler(r *application.Rules, j *application.Jobs, d *infrastructure.Store) *Scheduler {
	return &Scheduler{rules: r, jobs: j, db: d}
}
func (s *Scheduler) Tick(c context.Context) error {
	rs, e := s.rules.List(c)
	if e != nil {
		return e
	}
	now := time.Now()
	for _, r := range rs {
		if r.Enabled && !r.NextRun.After(now) {
			if _, e = s.jobs.Create(c, r, domain.JobArchive); e != nil {
				continue
			}
			r.Schedule(now)
			_ = s.db.SaveRule(c, r)
		}
	}
	return nil
}
