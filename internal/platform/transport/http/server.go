package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/9AliMay9/lyapus/internal/platform/config"
	"github.com/9AliMay9/lyapus/internal/platform/health"
)

func NewServer(cfg config.Config, logger *slog.Logger) *http.Server {
	mux := http.NewServeMux()
	healthHandler := health.NewHandler()

	mux.HandleFunc("/livez", healthHandler.Livez)
	mux.HandleFunc("/readyz", healthHandler.Readyz)

	return &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           requestLogger(logger, mux),
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)

		logger.Info(
			"http_request_completed",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start).String(),
		)
	})
}
