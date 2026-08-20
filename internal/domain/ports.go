package domain

import "context"

type RuleRepository interface {
	Save(context.Context, Rule) error
	Get(context.Context, string) (Rule, error)
	List(context.Context) ([]Rule, error)
	Delete(context.Context, string) error
}
type JobRepository interface {
	Save(context.Context, Job) error
	Get(context.Context, string) (Job, error)
	List(context.Context) ([]Job, error)
	Claim(context.Context, string, string) (Job, error)
	Events(context.Context, string) ([]Event, error)
	AddEvent(context.Context, Event) error
}
type BatchRepository interface {
	Save(context.Context, Batch) error
	Get(context.Context, string) (Batch, error)
	List(context.Context) ([]Batch, error)
}
type ArchiveStore interface {
	Create(ctx context.Context, rule Rule, jobID string, entries []ManifestEntry) (Batch, error)
	Verify(context.Context, Batch) (VerifyReport, error)
	Restore(context.Context, Batch, string, string) (RestoreReport, error)
}
type Event struct {
	ID      string `json:"id"`
	JobID   string `json:"job_id"`
	Kind    string `json:"kind"`
	Message string `json:"message"`
	At      string `json:"at"`
}
type VerifyReport struct {
	Files  int      `json:"files"`
	Good   int      `json:"good"`
	Bad    int      `json:"bad"`
	Errors []string `json:"errors,omitempty"`
}
type RestoreReport struct {
	Restored int      `json:"restored"`
	Skipped  int      `json:"skipped"`
	Failed   int      `json:"failed"`
	Errors   []string `json:"errors,omitempty"`
}
