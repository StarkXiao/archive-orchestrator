package infrastructure

import (
	"archive-orchestrator/internal/domain"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type State struct {
	Rules   []domain.Rule  `json:"rules"`
	Jobs    []domain.Job   `json:"jobs"`
	Batches []domain.Batch `json:"batches"`
	Events  []domain.Event `json:"events"`
}
type Store struct {
	mu    sync.RWMutex
	path  string
	state State
}

func NewStore(path string) (*Store, error) {
	s := &Store{path: path}
	if b, e := os.ReadFile(path); e == nil {
		if e = json.Unmarshal(b, &s.state); e != nil {
			return nil, e
		}
	}
	return s, nil
}
func (s *Store) flush() error {
	b, e := json.MarshalIndent(s.state, "", "  ")
	if e != nil {
		return e
	}
	if e = os.MkdirAll(filepath.Dir(s.path), 0755); e != nil {
		return e
	}
	tmp := s.path + ".tmp"
	if e = os.WriteFile(tmp, b, 0600); e != nil {
		return e
	}
	return os.Rename(tmp, s.path)
}
func (s *Store) Save(ctx context.Context, v any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch x := v.(type) {
	case domain.Rule:
		for i := range s.state.Rules {
			if s.state.Rules[i].ID == x.ID {
				s.state.Rules[i] = x
				return s.flush()
			}
		}
		s.state.Rules = append(s.state.Rules, x)
	case domain.Job:
		for i := range s.state.Jobs {
			if s.state.Jobs[i].ID == x.ID {
				s.state.Jobs[i] = x
				return s.flush()
			}
		}
		s.state.Jobs = append(s.state.Jobs, x)
	case domain.Batch:
		for i := range s.state.Batches {
			if s.state.Batches[i].ID == x.ID {
				s.state.Batches[i] = x
				return s.flush()
			}
		}
		s.state.Batches = append(s.state.Batches, x)
	}
	return s.flush()
}
func (s *Store) SaveRule(c context.Context, r domain.Rule) error   { return s.Save(c, r) }
func (s *Store) SaveJob(c context.Context, j domain.Job) error     { return s.Save(c, j) }
func (s *Store) SaveBatch(c context.Context, b domain.Batch) error { return s.Save(c, b) }
func (s *Store) GetRule(_ context.Context, id string) (domain.Rule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, v := range s.state.Rules {
		if v.ID == id {
			return v, nil
		}
	}
	return domain.Rule{}, domain.ErrNotFound
}
func (s *Store) ListRules(_ context.Context) ([]domain.Rule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]domain.Rule(nil), s.state.Rules...), nil
}
func (s *Store) DeleteRule(c context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, v := range s.state.Rules {
		if v.ID == id {
			s.state.Rules = append(s.state.Rules[:i], s.state.Rules[i+1:]...)
			return s.flush()
		}
	}
	return domain.ErrNotFound
}
func (s *Store) GetJob(_ context.Context, id string) (domain.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, v := range s.state.Jobs {
		if v.ID == id {
			return v, nil
		}
	}
	return domain.Job{}, domain.ErrNotFound
}
func (s *Store) ListJobs(_ context.Context) ([]domain.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]domain.Job(nil), s.state.Jobs...), nil
}
func (s *Store) Claim(c context.Context, id, worker string) (domain.Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.state.Jobs {
		j := &s.state.Jobs[i]
		if j.ID == id {
			if j.Status != domain.Pending && j.Status != domain.RetryWait {
				return domain.Job{}, domain.ErrConflict
			}
			now := time.Now()
			j.Status = domain.Running
			j.Attempts++
			j.StartedAt = &now
			lease := now.Add(2 * time.Minute)
			j.LeaseUntil = &lease
			j.LeaseOwner = worker
			if e := s.flush(); e != nil {
				return domain.Job{}, e
			}
			return *j, nil
		}
	}
	return domain.Job{}, domain.ErrNotFound
}
func (s *Store) RenewLease(id, worker string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.state.Jobs {
		j := &s.state.Jobs[i]
		if j.ID != id {
			continue
		}
		if j.Status != domain.Running || j.LeaseOwner != worker {
			return domain.ErrConflict
		}
		lease := now.Add(2 * time.Minute)
		j.LeaseUntil = &lease
		return s.flush()
	}
	return domain.ErrNotFound
}
func (s *Store) RequeueExpired(now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	changed := false
	for i := range s.state.Jobs {
		j := &s.state.Jobs[i]
		if j.Status == domain.Running && j.LeaseUntil != nil && j.LeaseUntil.Before(now) {
			j.Status = domain.RetryWait
			j.Error = "worker lease expired"
			j.LeaseUntil = nil
			retryAt := now
			j.RetryAt = &retryAt
			changed = true
		}
	}
	if !changed {
		return nil
	}
	return s.flush()
}
func (s *Store) Events(_ context.Context, id string) ([]domain.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.Event
	for _, e := range s.state.Events {
		if e.JobID == id {
			out = append(out, e)
		}
	}
	return out, nil
}
func (s *Store) AddEvent(c context.Context, e domain.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Events = append(s.state.Events, e)
	return s.flush()
}
func (s *Store) GetBatch(_ context.Context, id string) (domain.Batch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, v := range s.state.Batches {
		if v.ID == id {
			return v, nil
		}
	}
	return domain.Batch{}, domain.ErrNotFound
}
func (s *Store) ListBatches(_ context.Context) ([]domain.Batch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]domain.Batch(nil), s.state.Batches...), nil
}
