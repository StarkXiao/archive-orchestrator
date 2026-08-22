package worker

import (
	"archive-orchestrator/internal/application"
	"archive-orchestrator/internal/domain"
	"archive-orchestrator/internal/infrastructure"
	"context"
	"fmt"
	"time"
)

type Executor struct {
	jobs    *application.Jobs
	archive *application.Archive
	batches *application.Batches
	db      *infrastructure.Store
}

func NewExecutor(j *application.Jobs, a *application.Archive, b *application.Batches, d *infrastructure.Store) *Executor {
	return &Executor{jobs: j, archive: a, batches: b, db: d}
}
func (e *Executor) Run(c context.Context, j domain.Job) error {
	switch j.Type {
	case domain.JobArchive:
		b, err := e.archive.Run(c, j)
		if err != nil {
			return err
		}
		j.BatchID = b.ID
		if err := e.db.SaveJob(c, j); err != nil {
			return err
		}
		report, err := e.archive.Verify(c, b)
		if err != nil || report.Bad > 0 {
			if err != nil {
				return err
			}
			return fmt.Errorf("archive verification failed for %d files", report.Bad)
		}
		if err = e.archive.CleanSources(b); err != nil {
			return err
		}
		j.Files = b.Files
		j.Bytes = b.Bytes
		j.Status = domain.Succeeded
	case domain.JobVerify:
		b, err := e.batches.Get(c, j.RuleID)
		if err != nil {
			return err
		}
		r, err := e.archive.Verify(c, b)
		if err != nil {
			return err
		}
		if r.Bad > 0 {
			j.Status = domain.PartialFailed
			j.Error = fmt.Sprintf("%d files failed verification", r.Bad)
		} else {
			j.Status = domain.Succeeded
		}
	case domain.JobRestore:
		b, err := e.batches.Get(c, j.RuleID)
		if err != nil {
			return err
		}
		report, err := e.archive.Restore(c, b, j.RestoreTarget, j.ConflictMode)
		if err != nil {
			return err
		}
		j.Files = report.Restored + report.Skipped + report.Failed
		if report.Failed > 0 {
			j.Status = domain.PartialFailed
			j.Error = fmt.Sprintf("%d files failed to restore", report.Failed)
		} else {
			j.Status = domain.Succeeded
		}
	}
	now := time.Now()
	j.FinishedAt = &now
	return e.db.SaveJob(c, j)
}
