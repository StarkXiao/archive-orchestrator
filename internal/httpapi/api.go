package httpapi

import (
	"archive-orchestrator/internal/application"
	"archive-orchestrator/internal/domain"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type API struct {
	rules   *application.Rules
	jobs    *application.Jobs
	batches *application.Batches
}

func New(r *application.Rules, j *application.Jobs, b *application.Batches) *API {
	return &API{rules: r, jobs: j, batches: b}
}
func (a *API) Routes() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { write(w, 200, map[string]string{"status": "ok"}) })
	m.HandleFunc("/api/v1/rules", a.rulesHandler)
	m.HandleFunc("/api/v1/jobs", a.jobsHandler)
	m.HandleFunc("/api/v1/batches", a.batchHandler)
	m.HandleFunc("/api/v1/batches/", a.batchActionHandler)
	m.HandleFunc("/api/v1/jobs/", a.jobActionHandler)
	m.Handle("/", http.FileServer(http.Dir("web")))
	return m
}
func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func errStatus(e error) int {
	if errors.Is(e, domain.ErrNotFound) {
		return 404
	}
	if errors.Is(e, domain.ErrConflict) {
		return 409
	}
	return 400
}
func id(path, prefix string) string { return strings.TrimPrefix(path, prefix) }
