package cataloghttp

import (
	"errors"
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSONBody(t *testing.T) {
	request := httptest.NewRequest(
		stdhttp.MethodPost,
		"http://example.test/v1/teams",
		strings.NewReader(`{"name":"Platform"}`),
	)
	recorder := httptest.NewRecorder()

	var got struct {
		Name string `json:"name"`
	}
	if err := decodeJSONBody(recorder, request, &got); err != nil {
		t.Fatalf("decodeJSONBody() error = %v", err)
	}
	if got.Name != "Platform" {
		t.Fatalf("decoded name = %q, want %q", got.Name, "Platform")
	}
}

func TestDecodeJSONBodyRejectsUnknownFields(t *testing.T) {
	request := httptest.NewRequest(
		stdhttp.MethodPost,
		"http://example.test/v1/teams",
		strings.NewReader(`{"name":"Platform","unexpected":"true"}`),
	)
	recorder := httptest.NewRecorder()

	var destination struct {
		Name string `json:"name"`
	}
	err := decodeJSONBody(recorder, request, &destination)

	if err == nil {
		t.Fatal("decodeJSONBody() error = nil, want an error")
	}
	if !strings.Contains(err.Error(), `unknown field "unexpected"`) {
		t.Fatalf("decodeJSONBody() error = %q, want unknown field error", err)
	}
}

func TestDecodeJSONBodyRejectsMultipleValues(t *testing.T) {
	request := httptest.NewRequest(
		stdhttp.MethodPost,
		"http://example.test/v1/teams",
		strings.NewReader(`{"name":"Platform"} {"name":"Other"}`),
	)
	recorder := httptest.NewRecorder()

	var destination struct {
		Name string `json:"name"`
	}
	err := decodeJSONBody(recorder, request, &destination)

	if err == nil {
		t.Fatal("decodeJSONBody() error = nil, want an error")
	}
	if err.Error() != "request body must contain a single JSON value" {
		t.Fatalf("decodeJSONBody() error = %q, want single value error", err)
	}
}

func TestDecodeJSONBodyRejectsTooLargeBody(t *testing.T) {
	body := `{"name":"` + strings.Repeat("x", int(maxJSONRequestBodyBytes)) + `"}`

	request := httptest.NewRequest(
		stdhttp.MethodPost,
		"http://example.test/v1/teams",
		strings.NewReader(body),
	)
	recorder := httptest.NewRecorder()

	var destination struct {
		Name string `json:"name"`
	}
	err := decodeJSONBody(recorder, request, &destination)

	var maxBytesError *stdhttp.MaxBytesError
	if !errors.As(err, &maxBytesError) {
		t.Fatalf("decodeJSONBody() error = %v, want MaxBytesError", err)
	}
	if maxBytesError.Limit != maxJSONRequestBodyBytes {
		t.Fatalf("MaxBytesError limit = %d, want %d", maxBytesError.Limit, maxJSONRequestBodyBytes)
	}
}
