package domain

import "time"

type ManifestEntry struct {
	Source     string    `json:"source"`
	Archive    string    `json:"archive"`
	Size       int64     `json:"size"`
	SHA256     string    `json:"sha256"`
	ModifiedAt time.Time `json:"modified_at"`
}
type Batch struct {
	ID           string          `json:"id"`
	JobID        string          `json:"job_id"`
	Root         string          `json:"root"`
	Manifest     string          `json:"manifest"`
	Entries      []ManifestEntry `json:"entries"`
	Files        int             `json:"files"`
	Bytes        int64           `json:"bytes"`
	CreatedAt    time.Time       `json:"created_at"`
	LastVerify   *time.Time      `json:"last_verify,omitempty"`
	VerifyResult string          `json:"verify_result,omitempty"`
}
