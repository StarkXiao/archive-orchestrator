package domain

import "time"

type JobType string

const (
	JobArchive JobType = "archive"
	JobVerify  JobType = "verify"
	JobRestore JobType = "restore"
)

type JobStatus string

const (
	Pending       JobStatus = "pending"
	Running       JobStatus = "running"
	RetryWait     JobStatus = "retry_wait"
	Succeeded     JobStatus = "succeeded"
	PartialFailed JobStatus = "partial_failed"
	Failed        JobStatus = "failed"
	ManualReview  JobStatus = "manual_review"
	Cancelled     JobStatus = "cancelled"
)

type Job struct {
	ID            string     `json:"id"`
	RuleID        string     `json:"rule_id"`
	Type          JobType    `json:"type"`
	Status        JobStatus  `json:"status"`
	RuleSnapshot  Rule       `json:"rule_snapshot"`
	ScheduledAt   time.Time  `json:"scheduled_at"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
	Attempts      int        `json:"attempts"`
	Files         int        `json:"files"`
	Bytes         int64      `json:"bytes"`
	Error         string     `json:"error,omitempty"`
	RetryAt       *time.Time `json:"retry_at,omitempty"`
	RestoreTarget string     `json:"restore_target,omitempty"`
	ConflictMode  string     `json:"conflict_mode,omitempty"`
	BatchID       string     `json:"batch_id,omitempty"`
	LeaseUntil    *time.Time `json:"lease_until,omitempty"`
	LeaseOwner    string     `json:"lease_owner,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

func (j Job) Terminal() bool {
	return j.Status == Succeeded || j.Status == PartialFailed || j.Status == Failed || j.Status == ManualReview || j.Status == Cancelled
}
func (j *Job) Start(now time.Time) error {
	if j.Terminal() || j.Status == Running {
		return ErrConflict
	}
	j.Status = Running
	j.Attempts++
	j.StartedAt = &now
	return nil
}
func (j *Job) Finish(status JobStatus, err error, now time.Time) {
	j.Status = status
	j.FinishedAt = &now
	if err != nil {
		j.Error = err.Error()
	}
}
