package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/9AliMay9/lyapus/internal/platform/config"
	"github.com/9AliMay9/lyapus/internal/platform/health"
	"github.com/9AliMay9/lyapus/internal/platform/requestid"
)

func NewServer(
	cfg config.Config,
	logger *slog.Logger,
	pinger health.Pinger,
	catalogHandler http.Handler,
) *http.Server {
	mux := http.NewServeMux()
	healthHandler := health.NewHandler(pinger)

	mux.HandleFunc("/livez", healthHandler.Livez)
	mux.HandleFunc("/readyz", healthHandler.Readyz)
	mux.Handle("/v1/", catalogHandler)

	return &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           requestLogger(logger, requestid.Middleware(mux)),
		ReadHeaderTimeout: 5 * time.Second,
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *statusRecorder) WriteHeader(status int) {
	if w.status != 0 {
		return
	}

	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusRecorder) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}

	return w.ResponseWriter.Write(body)
}

func (w *statusRecorder) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{
			ResponseWriter: w,
		}

		next.ServeHTTP(recorder, r)

		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}

		logger.Info(
			"http_request_completed",
			"request_id", recorder.Header().Get("X-Request-ID"),
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
			"duration", time.Since(start).String(),
		)
	})
}
