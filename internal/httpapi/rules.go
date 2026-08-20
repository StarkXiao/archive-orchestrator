package httpapi

import (
	"archive-orchestrator/internal/domain"
	"encoding/json"
	"net/http"
)

func (a *API) rulesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		v, e := a.rules.List(r.Context())
		if e != nil {
			write(w, 500, map[string]string{"error": e.Error()})
			return
		}
		write(w, 200, v)
		return
	}
	if r.Method == "POST" {
		var v domain.Rule
		if json.NewDecoder(r.Body).Decode(&v) != nil {
			write(w, 400, map[string]string{"error": "invalid json"})
			return
		}
		v, e := a.rules.Create(r.Context(), v)
		if e != nil {
			write(w, errStatus(e), map[string]string{"error": e.Error()})
			return
		}
		write(w, 201, v)
		return
	}
	write(w, 405, nil)
}
