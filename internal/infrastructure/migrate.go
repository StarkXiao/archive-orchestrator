package infrastructure

import (
	"context"
	"time"
)

type Migrator struct{ store *Store }

func NewMigrator(s *Store) *Migrator              { return &Migrator{store: s} }
func (m *Migrator) Run(ctx context.Context) error { _, e := m.store.ListRules(ctx); return e }
func NewID(prefix string) string                  { return prefix + "-" + time.Now().Format("20060102150405.000000000") }
