package infrastructure

import (
	"archive-orchestrator/internal/domain"
	"context"
	"time"
)

type Audit struct{ store *Store }

func NewAudit(s *Store) *Audit { return &Audit{store: s} }
func (a *Audit) Record(c context.Context, job, kind, msg string) error {
	return a.store.AddEvent(c, domain.Event{ID: NewID("event"), JobID: job, Kind: kind, Message: msg, At: time.Now().UTC().Format(time.RFC3339)})
}
