package requestid

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewareGeneratesAndStoresRequestID(t *testing.T) {
	const generatedID = "server-generated-request-id"

	request := httptest.NewRequest(
		http.MethodGet,
		"http://example.test/livez",
		nil,
	)
	request.Header.Set(Header, "client-supplied-request-id")

	recorder := httptest.NewRecorder()
	handler := middlewareWithGenerator(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := FromContext(r.Context()); got != generatedID {
				t.Fatalf("request ID in context = %q, want %q", got, generatedID)
			}
			w.WriteHeader(http.StatusNoContent)
		}),
		func() string {
			return generatedID
		},
	)

	handler.ServeHTTP(recorder, request)

	if got := recorder.Header().Get(Header); got != generatedID {
		t.Fatalf("response request ID = %q, want %q", got, generatedID)
	}
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}
