package httpapi

import (
	"log/slog"
	"net/http"
	"time"
)

func Logging(next http.Handler, l *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		l.Info("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start).String())
	})
}
