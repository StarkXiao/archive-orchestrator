package application

import (
	"archive-orchestrator/internal/domain"
	"archive-orchestrator/internal/infrastructure"
	"context"
	"path/filepath"
)

type Batches struct {
	db          *infrastructure.Store
	files       *infrastructure.Files
	jobs        *Jobs
	restoreRoot string
}

func NewBatches(db *infrastructure.Store, f *infrastructure.Files, j *Jobs, restoreRoot string) *Batches {
	return &Batches{db: db, files: f, jobs: j, restoreRoot: restoreRoot}
}
func (b *Batches) List(c context.Context) ([]domain.Batch, error) { return b.db.ListBatches(c) }
func (b *Batches) Get(c context.Context, id string) (domain.Batch, error) {
	return b.db.GetBatch(c, id)
}
func (b *Batches) Verify(c context.Context, id string) (domain.Job, error) {
	v, e := b.Get(c, id)
	if e != nil {
		return domain.Job{}, e
	}
	r, e := b.jobs.Create(c, domain.Rule{ID: v.ID, Name: "batch-verify", MaxAttempts: 3}, domain.JobVerify)
	return r, e
}
func (b *Batches) Restore(c context.Context, id, target, conflict string) (domain.Job, error) {
	v, e := b.Get(c, id)
	if e != nil {
		return domain.Job{}, e
	}
	root, e := filepath.Abs(b.restoreRoot)
	if e != nil {
		return domain.Job{}, e
	}
	destination, e := filepath.Abs(target)
	if e != nil {
		return domain.Job{}, e
	}
	rel, e := filepath.Rel(root, destination)
	if e != nil || rel == "." {
		return domain.Job{}, domain.ErrInvalid
	}
	if conflict != "skip" && conflict != "overwrite" {
		return domain.Job{}, domain.ErrInvalid
	}
	r, e := b.jobs.Create(c, domain.Rule{ID: v.ID, Name: destination, ArchiveRoot: v.Root, MaxAttempts: 3}, domain.JobRestore)
	if e != nil {
		return r, e
	}
	r.RestoreTarget = destination
	r.ConflictMode = conflict
	return r, b.db.SaveJob(c, r)
}
