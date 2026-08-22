package application

import (
	"archive-orchestrator/internal/domain"
	"archive-orchestrator/internal/infrastructure"
	"context"
	"strings"
	"time"
)

type Rules struct{ db *infrastructure.Store }

func NewRules(db *infrastructure.Store) *Rules { return &Rules{db: db} }
func (r *Rules) Create(c context.Context, v domain.Rule) (domain.Rule, error) {
	if e := v.Validate(); e != nil {
		return v, e
	}
	v.ID = infrastructure.NewID("rule")
	v.Version = 1
	v.CreatedAt = time.Now()
	v.UpdatedAt = v.CreatedAt
	v.Schedule(v.CreatedAt)
	return v, r.db.SaveRule(c, v)
}
func (r *Rules) Get(c context.Context, id string) (domain.Rule, error) { return r.db.GetRule(c, id) }
func (r *Rules) List(c context.Context) ([]domain.Rule, error) {
	v, e := r.db.ListRules(c)
	if e != nil {
		return nil, e
	}
	for i := range v {
		v[i].Name = strings.TrimSpace(v[i].Name)
	}
	return v, nil
}
func (r *Rules) Update(c context.Context, v domain.Rule) (domain.Rule, error) {
	old, e := r.db.GetRule(c, v.ID)
	if e != nil {
		return v, e
	}
	if old.Version != v.Version {
		return v, domain.ErrConflict
	}
	if e = v.Validate(); e != nil {
		return v, e
	}
	v.Version++
	v.UpdatedAt = time.Now()
	return v, r.db.SaveRule(c, v)
}
func (r *Rules) SetEnabled(c context.Context, id string, on bool) (domain.Rule, error) {
	v, e := r.Get(c, id)
	if e != nil {
		return v, e
	}
	v.Enabled = on
	v.Version++
	v.UpdatedAt = time.Now()
	return v, r.db.SaveRule(c, v)
}
