package health

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerHealthChecks(t *testing.T) {
	handler := NewHandler()

	tests := []struct {
		name    string
		handler http.HandlerFunc
	}{
		{
			name:    "livez",
			handler: handler.Livez,
		},
		{
			name:    "readyz",
			handler: handler.Readyz,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/"+tt.name, nil)
			recorder := httptest.NewRecorder()

			tt.handler(recorder, req)

			response := recorder.Result()
			defer response.Body.Close()

			if response.StatusCode != http.StatusOK {
				t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
			}
			if contentType := response.Header.Get("Content-Type"); contentType != "application/json" {
				t.Fatalf("Content-Type = %q, want %q", contentType, "application/json")
			}

			body := recorder.Body.String()
			if body != "{\"status\":\"ok\"}\n" {
				t.Fatalf("body = %q, want %q", body, "{\"status\":\"ok\"}\n")
			}
		})
	}
}

func TestHandlerRejectsNonGETRequests(t *testing.T) {
	handler := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/livez", nil)
	recorder := httptest.NewRecorder()

	handler.Livez(recorder, req)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
	}
}
