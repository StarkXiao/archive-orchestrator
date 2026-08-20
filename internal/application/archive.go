package application

import (
	"archive-orchestrator/internal/domain"
	"archive-orchestrator/internal/infrastructure"
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Archive struct {
	db    *infrastructure.Store
	files *infrastructure.Files
	jobs  *Jobs
}

func NewArchive(db *infrastructure.Store, f *infrastructure.Files, j *Jobs) *Archive {
	return &Archive{db: db, files: f, jobs: j}
}
func (a *Archive) Run(c context.Context, j domain.Job) (domain.Batch, error) {
	if j.BatchID != "" {
		return a.db.GetBatch(c, j.BatchID)
	}
	var entries []domain.ManifestEntry
	cut := time.Now().Add(-time.Duration(j.RuleSnapshot.RetentionDays) * 24 * time.Hour)
	e := filepath.Walk(j.RuleSnapshot.SourceRoot, func(p string, i os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if i.IsDir() || !i.Mode().IsRegular() {
			return nil
		}
		ok, _ := filepath.Match(j.RuleSnapshot.Pattern, filepath.Base(p))
		if !ok || i.ModTime().After(cut) {
			return nil
		}
		entries = append(entries, domain.ManifestEntry{Source: p, ModifiedAt: i.ModTime()})
		if j.RuleSnapshot.MaxFiles > 0 && len(entries) >= j.RuleSnapshot.MaxFiles {
			return filepath.SkipDir
		}
		return nil
	})
	if e != nil {
		return domain.Batch{}, e
	}
	return a.files.Create(c, j.RuleSnapshot, j.ID, entries)
}
func (a *Archive) CleanSources(b domain.Batch) error {
	for _, x := range b.Entries {
		if e := os.Remove(x.Source); e != nil && !strings.Contains(e.Error(), "no such file") {
			return e
		}
	}
	return nil
}

func (a *Archive) Verify(ctx context.Context, b domain.Batch) (domain.VerifyReport, error) {
	return a.files.Verify(ctx, b)
}

func (a *Archive) Restore(ctx context.Context, b domain.Batch, target, conflict string) (domain.RestoreReport, error) {
	return a.files.Restore(ctx, b, target, conflict)
}
