package cataloghttp

import (
	"encoding/json"
	"errors"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/9AliMay9/lyapus/internal/catalog"
)

func TestWriteJSON(t *testing.T) {
	recorder := httptest.NewRecorder()

	writeJSON(recorder, stdhttp.StatusCreated, struct {
		Status string `json:"status"`
	}{
		Status: "created",
	})

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != stdhttp.StatusCreated {
		t.Fatalf("status = %d, want %d", response.StatusCode, stdhttp.StatusCreated)
	}
	if contentType := response.Header.Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", contentType, "application/json")
	}

	var got struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if got.Status != "created" {
		t.Fatalf("status body = %q, want %q", got.Status, "created")
	}
}

func TestWriteError(t *testing.T) {
	recorder := httptest.NewRecorder()

	writeError(
		recorder,
		stdhttp.StatusBadRequest,
		"invalid_argument",
		"name must not be empty",
		"request-123",
	)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != stdhttp.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.StatusCode, stdhttp.StatusBadRequest)
	}

	var got errorResponse
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	want := errorResponse{
		Code:      "invalid_argument",
		Message:   "name must not be empty",
		RequestID: "request-123",
	}
	if got != want {
		t.Fatalf("error body = %#v, want %#v", got, want)
	}
}

func TestWriteCatalogError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   errorResponse
	}{
		{
			name:       "invalid argument with public message",
			err:        &catalog.InvalidArgumentError{Message: "name must not be empty"},
			wantStatus: stdhttp.StatusBadRequest,
			wantBody: errorResponse{
				Code:      "invalid_argument",
				Message:   "name must not be empty",
				RequestID: "request-123",
			},
		},
		{
			name:       "not found",
			err:        catalog.ErrNotFound,
			wantStatus: stdhttp.StatusNotFound,
			wantBody: errorResponse{
				Code:      "not_found",
				Message:   "resource not found",
				RequestID: "request-123",
			},
		},
		{
			name:       "conflict",
			err:        catalog.ErrConflict,
			wantStatus: stdhttp.StatusConflict,
			wantBody: errorResponse{
				Code:      "conflict",
				Message:   "resource conflict",
				RequestID: "request-123",
			},
		},
		{
			name:       "unknown error is not exposed",
			err:        errors.New("postgres://user:secret@database:5432/catalog"),
			wantStatus: stdhttp.StatusInternalServerError,
			wantBody: errorResponse{
				Code:      "internal",
				Message:   "internal server error",
				RequestID: "request-123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			writeCatalogError(recorder, tt.err, "request-123")

			response := recorder.Result()
			defer response.Body.Close()

			if response.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.StatusCode, tt.wantStatus)
			}

			var got errorResponse
			if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
				t.Fatalf("decode response body: %v", err)
			}
			if got != tt.wantBody {
				t.Fatalf("error body = %#v, want %#v", got, tt.wantBody)
			}
		})
	}
}
