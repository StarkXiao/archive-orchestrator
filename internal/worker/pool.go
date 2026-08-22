package worker

import (
	"archive-orchestrator/internal/application"
	"archive-orchestrator/internal/domain"
	"archive-orchestrator/internal/infrastructure"
	"context"
	"fmt"
	"time"
)

type Pool struct {
	jobs    *application.Jobs
	exec    *Executor
	db      *infrastructure.Store
	workers int
}

func NewPool(j *application.Jobs, e *Executor, d *infrastructure.Store, n int) *Pool {
	return &Pool{jobs: j, exec: e, db: d, workers: n}
}
func (p *Pool) Run(c context.Context) {
	for i := 0; i < p.workers; i++ {
		go p.loop(c, fmt.Sprintf("worker-%d", i))
	}
}
func (p *Pool) loop(c context.Context, workerID string) {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-c.Done():
			return
		case now := <-t.C:
			_ = p.db.RequeueExpired(now)
			js, _ := p.jobs.List(c)
			for _, j := range js {
				if j.Status == domain.Pending || (j.Status == domain.RetryWait && (j.RetryAt == nil || !j.RetryAt.After(now))) {
					if claimed, e := p.db.Claim(c, j.ID, workerID); e == nil {
						if e = p.runClaimed(c, claimed); e != nil {
							latest, getErr := p.jobs.Get(c, claimed.ID)
							if getErr == nil {
								_ = p.jobs.Fail(c, latest, e)
							}
						}
					}
				}
			}
		}
	}
}

func (p *Pool) runClaimed(ctx context.Context, job domain.Job) error {
	done := make(chan struct{})
	defer close(done)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case now := <-ticker.C:
				_ = p.db.RenewLease(job.ID, "lease-renewer", now)
			}
		}
	}()
	return p.exec.Run(ctx, job)
}
