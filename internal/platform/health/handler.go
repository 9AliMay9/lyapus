package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

const readinessTimeout = time.Second

type Pinger interface {
	Ping(context.Context) error
}

type Handler struct {
	pinger Pinger
}

func NewHandler(pinger Pinger) Handler {
	return Handler{pinger: pinger}
}

func (Handler) Livez(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}

	respondOK(w)
}

func (h Handler) Readyz(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), readinessTimeout)
	defer cancel()

	if err := h.pinger.Ping(ctx); err != nil {
		respondNotReady(w)
		return
	}

	respondOK(w)
}

func requireGET(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodGet {
		return true
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
	return false
}

func respondOK(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func respondNotReady(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "not_ready"})
}
