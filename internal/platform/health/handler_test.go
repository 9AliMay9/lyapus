package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type pingerFunc func(context.Context) error

func (f pingerFunc) Ping(ctx context.Context) error {
	return f(ctx)
}

func TestHandlerLivezDoesNotCheckDatabase(t *testing.T) {
	handler := NewHandler(pingerFunc(func(context.Context) error {
		t.Fatal("Ping() must not be called by /livez")
		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/livez", nil)
	recorder := httptest.NewRecorder()

	handler.Livez(recorder, req)

	assertStatusAndBody(t, recorder, http.StatusOK, "{\"status\":\"ok\"}\n")
}

func TestHandlerReadyzReturnsOKWhenDatabaseIsReachable(t *testing.T) {
	pingCalled := false
	handler := NewHandler(pingerFunc(func(ctx context.Context) error {
		pingCalled = true

		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("Ping() context has no deadline")
		}

		return nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	recorder := httptest.NewRecorder()

	handler.Readyz(recorder, req)

	if !pingCalled {
		t.Fatal("Ping() was not called")
	}
	assertStatusAndBody(t, recorder, http.StatusOK, "{\"status\":\"ok\"}\n")
}

func TestHandlerReadyzReturnsServiceUnavailableWhenDatabaseIsUnreachable(t *testing.T) {
	handler := NewHandler(pingerFunc(func(context.Context) error {
		return errors.New("database is unavailable")
	}))

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	recorder := httptest.NewRecorder()

	handler.Readyz(recorder, req)

	assertStatusAndBody(t, recorder, http.StatusServiceUnavailable, "{\"status\":\"not_ready\"}\n")
}

func TestHandlerRejectsNonGETRequests(t *testing.T) {
	handler := NewHandler(pingerFunc(func(context.Context) error {
		t.Fatal("Ping() must not be called for a non-GET request")
		return nil
	}))

	req := httptest.NewRequest(http.MethodPost, "/livez", nil)
	recorder := httptest.NewRecorder()

	handler.Livez(recorder, req)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
	}
}

func assertStatusAndBody(t *testing.T, recorder *httptest.ResponseRecorder, wantStatus int, wantBody string) {
	t.Helper()

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != wantStatus {
		t.Fatalf("status = %d, want %d", response.StatusCode, wantStatus)
	}
	if contentType := response.Header.Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", contentType, "application/json")
	}
	if body := recorder.Body.String(); body != wantBody {
		t.Fatalf("body = %q, want %q", body, wantBody)
	}
}
