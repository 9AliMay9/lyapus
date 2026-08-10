package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/9AliMay9/lyapus/internal/platform/config"
	"github.com/9AliMay9/lyapus/internal/platform/requestid"
)

type successfulPinger struct{}

func (successfulPinger) Ping(context.Context) error {
	return nil
}

func TestNewServerRegistersHealthChecks(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	server := NewServer(
		config.Config{HTTPAddr: "127.0.0.1:8080"},
		logger,
		successfulPinger{},
		http.NotFoundHandler(),
	)

	if server.Addr != "127.0.0.1:8080" {
		t.Fatalf("Addr = %q, want %q", server.Addr, "127.0.0.1:8080")
	}

	for _, path := range []string{"/livez", "/readyz"} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			recorder := httptest.NewRecorder()

			server.Handler.ServeHTTP(recorder, req)

			if requestID := recorder.Header().Get(requestid.Header); requestID == "" {
				t.Fatalf("%s response header = empty, want generated ID", requestid.Header)
			}

			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
			}
			if body := recorder.Body.String(); body != "{\"status\":\"ok\"}\n" {
				t.Fatalf("body = %q, want %q", body, "{\"status\":\"ok\"}\n")
			}
		})
	}
}

func TestNewServerMountsCatalogHandler(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	catalogHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/teams" {
			t.Fatalf("path = %q, want %q", r.URL.Path, "/v1/teams")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	server := NewServer(
		config.Config{HTTPAddr: "127.0.0.1:8080"},
		logger,
		successfulPinger{},
		catalogHandler,
	)

	request := httptest.NewRequest(http.MethodPost, "/v1/teams", nil)
	recorder := httptest.NewRecorder()

	server.Handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}

func TestRequestLoggerIncludesRequestIDAndStatus(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))

	handler := requestLogger(
		logger,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Request-ID", "request-123")
			w.WriteHeader(http.StatusCreated)
		}),
	)

	request := httptest.NewRequest(http.MethodPost, "/v1/teams", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("decode log entry: %v", err)
	}

	if got := entry["request_id"]; got != "request-123" {
		t.Fatalf("request_id = %#v, want %q", got, "request-123")
	}
	if got := entry["method"]; got != http.MethodPost {
		t.Fatalf("method = %#v, want %q", got, http.MethodPost)
	}
	if got := entry["path"]; got != "/v1/teams" {
		t.Fatalf("path = %#v, want %q", got, "/v1/teams")
	}
	if got := entry["status"]; got != float64(http.StatusCreated) {
		t.Fatalf("status = %#v, want %d", got, http.StatusCreated)
	}
}
