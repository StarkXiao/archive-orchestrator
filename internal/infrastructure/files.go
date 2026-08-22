package infrastructure

import (
	"archive-orchestrator/internal/domain"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Files struct{ batches *Store }

func NewFiles(s *Store) *Files { return &Files{batches: s} }
func digest(path string) (string, int64, error) {
	f, e := os.Open(path)
	if e != nil {
		return "", 0, e
	}
	defer f.Close()
	h := sha256.New()
	n, e := io.Copy(h, f)
	return hex.EncodeToString(h.Sum(nil)), n, e
}
func (f *Files) Create(ctx context.Context, r domain.Rule, jobID string, entries []domain.ManifestEntry) (domain.Batch, error) {
	id := fmt.Sprintf("batch-%d", time.Now().UnixNano())
	root := filepath.Join(r.ArchiveRoot, id)
	tempRoot := root + ".tmp"
	if e := os.MkdirAll(tempRoot, 0755); e != nil {
		return domain.Batch{}, e
	}
	defer os.RemoveAll(tempRoot)
	for i := range entries {
		src := entries[i].Source
		rel, e := filepath.Rel(r.SourceRoot, src)
		if e != nil || rel == "." || strings.HasPrefix(rel, "..") {
			return domain.Batch{}, fmt.Errorf("invalid archive source: %s", src)
		}
		dst := filepath.Join(tempRoot, rel)
		if e := os.MkdirAll(filepath.Dir(dst), 0755); e != nil {
			return domain.Batch{}, e
		}
		if e := copyFile(src, dst); e != nil {
			return domain.Batch{}, e
		}
		sum, n, e := digest(dst)
		if e != nil {
			return domain.Batch{}, e
		}
		entries[i].Archive = dst
		entries[i].SHA256 = sum
		entries[i].Size = n
	}
	manifest := filepath.Join(tempRoot, "manifest.json")
	b, _ := json.MarshalIndent(entries, "", "  ")
	if e := os.WriteFile(manifest, b, 0600); e != nil {
		return domain.Batch{}, e
	}
	if e := os.Rename(tempRoot, root); e != nil {
		return domain.Batch{}, e
	}
	for i := range entries {
		entries[i].Archive = strings.Replace(entries[i].Archive, tempRoot, root, 1)
	}
	manifest = filepath.Join(root, "manifest.json")
	b, e := json.MarshalIndent(entries, "", "  ")
	if e != nil {
		return domain.Batch{}, e
	}
	if e := os.WriteFile(manifest, b, 0600); e != nil {
		return domain.Batch{}, e
	}
	batch := domain.Batch{ID: id, JobID: jobID, Root: root, Manifest: manifest, Entries: entries, Files: len(entries), CreatedAt: time.Now()}
	for _, x := range entries {
		batch.Bytes += x.Size
	}
	return batch, f.batches.SaveBatch(ctx, batch)
}
func copyFile(a, b string) error {
	in, e := os.Open(a)
	if e != nil {
		return e
	}
	defer in.Close()
	out, e := os.Create(b)
	if e != nil {
		return e
	}
	// Close the destination before deciding whether to keep it; a successful
	// result must not be cleaned up, and the handle must be closed regardless.
	closeAndMaybeRemove := func(keep bool) {
		_ = out.Close()
		if !keep {
			_ = os.Remove(b)
		}
	}
	if _, e := io.Copy(out, in); e != nil {
		closeAndMaybeRemove(false)
		return e
	}
	closeAndMaybeRemove(true)
	return nil
}
func (f *Files) Verify(_ context.Context, b domain.Batch) (domain.VerifyReport, error) {
	r := domain.VerifyReport{Files: len(b.Entries)}
	for _, x := range b.Entries {
		sum, n, e := digest(x.Archive)
		if e != nil || n != x.Size || sum != x.SHA256 {
			r.Bad++
			r.Errors = append(r.Errors, x.Source)
		} else {
			r.Good++
		}
	}
	return r, nil
}
func (f *Files) Restore(_ context.Context, b domain.Batch, target, conflict string) (domain.RestoreReport, error) {
	r := domain.RestoreReport{}
	if e := os.MkdirAll(target, 0755); e != nil {
		return r, e
	}
	for _, x := range b.Entries {
		rel, e := filepath.Rel(b.Root, x.Archive)
		if e != nil || rel == "." || strings.HasPrefix(rel, "..") {
			r.Failed++
			r.Errors = append(r.Errors, x.Source)
			continue
		}
		dst := filepath.Join(target, rel)
		if _, e := os.Stat(dst); e == nil && conflict != "overwrite" {
			r.Skipped++
			continue
		}
		if e := os.MkdirAll(filepath.Dir(dst), 0755); e != nil {
			r.Failed++
			r.Errors = append(r.Errors, x.Source)
			continue
		}
		temp := dst + ".restore-tmp"
		if e := copyFile(x.Archive, temp); e != nil {
			r.Failed++
			r.Errors = append(r.Errors, x.Source)
			_ = os.Remove(temp)
		} else if sum, n, e := digest(temp); e != nil || sum != x.SHA256 || n != x.Size {
			r.Failed++
			r.Errors = append(r.Errors, x.Source)
			_ = os.Remove(temp)
		} else if e := os.Rename(temp, dst); e != nil {
			r.Failed++
			r.Errors = append(r.Errors, x.Source)
			_ = os.Remove(temp)
		} else {
			r.Restored++
		}
	}
	return r, nil
}
