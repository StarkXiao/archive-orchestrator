package httpapi

import (
	"net/http"
	"strings"
)

func (a *API) jobsHandler(w http.ResponseWriter, r *http.Request) {
	v, e := a.jobs.List(r.Context())
	if e != nil {
		write(w, 500, map[string]string{"error": e.Error()})
		return
	}
	write(w, 200, v)
}

func (a *API) jobActionHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/jobs/"), "/")
	if len(parts) != 2 || parts[0] == "" || r.Method != http.MethodPost {
		write(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	var result any
	var err error
	switch parts[1] {
	case "retry":
		result, err = a.jobs.Retry(r.Context(), parts[0])
	case "cancel":
		result, err = a.jobs.Cancel(r.Context(), parts[0])
	default:
		write(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	if err != nil {
		write(w, errStatus(err), map[string]string{"error": err.Error()})
		return
	}
	write(w, http.StatusOK, result)
}
