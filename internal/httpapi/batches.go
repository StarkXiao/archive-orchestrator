package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (a *API) batchHandler(w http.ResponseWriter, r *http.Request) {
	v, e := a.batches.List(r.Context())
	if e != nil {
		write(w, 500, map[string]string{"error": e.Error()})
		return
	}
	write(w, 200, v)
}

func (a *API) batchActionHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/batches/"), "/")
	if len(parts) != 2 || parts[0] == "" || r.Method != http.MethodPost {
		write(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	var result any
	var err error
	switch parts[1] {
	case "verify":
		result, err = a.batches.Verify(r.Context(), parts[0])
	case "restore":
		var request struct {
			Target   string `json:"target"`
			Conflict string `json:"conflict"`
		}
		if json.NewDecoder(r.Body).Decode(&request) != nil {
			write(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		result, err = a.batches.Restore(r.Context(), parts[0], request.Target, request.Conflict)
	default:
		write(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	if err != nil {
		write(w, errStatus(err), map[string]string{"error": err.Error()})
		return
	}
	write(w, http.StatusAccepted, result)
}
