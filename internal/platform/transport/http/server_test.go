package http

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/9Alimay/lyapus/internal/platform/config"
)

func TestNewServerRegistersHealthChecks(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	server := NewServer(config.Config{HTTPAddr: "127.0.0.1:8080"}, logger)

	if server.Addr != "127.0.0.1:8080" {
		t.Fatalf("Addr = %q, want %q", server.Addr, "127.0.0.1:8080")
	}

	for _, path := range []string{"/livez", "/readyz"} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			recorder := httptest.NewRecorder()

			server.Handler.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
			}
			if body := recorder.Body.String(); body != "{\"status\":\"ok\"}\n" {
				t.Fatalf("body = %q, want %q", body, "{\"status\":\"ok\"}\n")
			}
		})
	}
}
