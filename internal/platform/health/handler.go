package health

import (
	"encoding/json"
	"net/http"
)

type Handler struct{}

func NewHandler() Handler {
	return Handler{}
}

func (Handler) Livez(w http.ResponseWriter, r *http.Request) {
	respondOK(w, r)
}

func (Handler) Readyz(w http.ResponseWriter, r *http.Request) {
	respondOK(w, r)
}

func respondOK(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
