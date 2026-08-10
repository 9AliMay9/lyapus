package cataloghttp

import (
	"context"
	"encoding/json"
	"errors"
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/9AliMay9/lyapus/internal/catalog"
	"github.com/9AliMay9/lyapus/internal/platform/requestid"
)

func TestHandlerCreateTeam(t *testing.T) {
	createdAt := time.Date(2026, time.August, 9, 12, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	service := &recordingTeamService{
		team: catalog.Team{
			ID:        42,
			Slug:      "platform",
			Name:      "Platform",
			CreatedAt: createdAt,
			UpdatedAt: createdAt.Add(time.Hour),
		},
	}
	handler := newTestHandler(service)

	request := httptest.NewRequest(
		stdhttp.MethodPost,
		"/v1/teams",
		strings.NewReader(`{"slug":"platform","name":"Platform"}`),
	)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != stdhttp.StatusCreated {
		t.Fatalf("status = %d, want %d", response.StatusCode, stdhttp.StatusCreated)
	}
	if requestID := response.Header.Get(requestid.Header); requestID == "" {
		t.Fatalf("%s response header = empty, want generated ID", requestid.Header)
	}
	if service.createInput == nil {
		t.Fatal("CreateTeam() was not called")
	}

	wantInput := catalog.CreateTeamInput{
		Slug: "platform",
		Name: "Platform",
	}
	if *service.createInput != wantInput {
		t.Fatalf("CreateTeam() input = %#v, want %#v", *service.createInput, wantInput)
	}

	var got teamResponse
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if got.ID != 42 {
		t.Fatalf("response ID = %d, want 42", got.ID)
	}
	if got.Slug != "platform" {
		t.Fatalf("response slug = %q, want %q", got.Slug, "platform")
	}
	if got.Name != "Platform" {
		t.Fatalf("response name = %q, want %q", got.Name, "Platform")
	}
	if !got.CreatedAt.Equal(createdAt.UTC()) {
		t.Fatalf("response created_at = %s, want %s", got.CreatedAt, createdAt.UTC())
	}
	if got.CreatedAt.Location() != time.UTC {
		t.Fatalf("response created_at location = %s, want UTC", got.CreatedAt.Location())
	}
	if !got.UpdatedAt.Equal(createdAt.Add(time.Hour).UTC()) {
		t.Fatalf("response updated_at = %s, want %s", got.UpdatedAt, createdAt.Add(time.Hour).UTC())
	}
	if got.UpdatedAt.Location() != time.UTC {
		t.Fatalf("response updated_at location = %s, want UTC", got.UpdatedAt.Location())
	}
}

func TestHandlerCreateTeamRejectsInvalidRequestBody(t *testing.T) {
	service := &recordingTeamService{}
	handler := newTestHandler(service)

	request := httptest.NewRequest(
		stdhttp.MethodPost,
		"/v1/teams",
		strings.NewReader(`{"slug":"platform","name":"Platform","unexpected":"true"}`),
	)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if service.createInput != nil {
		t.Fatalf("CreateTeam() input = %#v, want no call", *service.createInput)
	}
	assertErrorResponse(
		t,
		recorder,
		stdhttp.StatusBadRequest,
		"invalid_argument",
		"invalid request body",
	)
}

func TestHandlerCreateTeamMapsCatalogErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantStatus  int
		wantCode    string
		wantMessage string
	}{
		{
			name:        "invalid argument",
			err:         &catalog.InvalidArgumentError{Message: "slug must match required format"},
			wantStatus:  stdhttp.StatusBadRequest,
			wantCode:    "invalid_argument",
			wantMessage: "slug must match required format",
		},
		{
			name:        "conflict",
			err:         catalog.ErrConflict,
			wantStatus:  stdhttp.StatusConflict,
			wantCode:    "conflict",
			wantMessage: "resource conflict",
		},
		{
			name:        "unknown error",
			err:         errors.New("postgres://user:secret@database:5432/catalog"),
			wantStatus:  stdhttp.StatusInternalServerError,
			wantCode:    "internal",
			wantMessage: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &recordingTeamService{
				err: tt.err,
			}
			handler := newTestHandler(service)

			request := httptest.NewRequest(
				stdhttp.MethodPost,
				"/v1/teams",
				strings.NewReader(`{"slug":"platform","name":"Platform"}`),
			)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			assertErrorResponse(
				t,
				recorder,
				tt.wantStatus,
				tt.wantCode,
				tt.wantMessage,
			)
		})
	}
}

func TestHandlerWritesStructuredNotFoundAndMethodNotAllowed(t *testing.T) {
	handler := newTestHandler(&recordingTeamService{})

	tests := []struct {
		name        string
		request     *stdhttp.Request
		wantStatus  int
		wantCode    string
		wantMessage string
	}{
		{
			name:        "not found",
			request:     httptest.NewRequest(stdhttp.MethodGet, "/v1/missing", nil),
			wantStatus:  stdhttp.StatusNotFound,
			wantCode:    "not_found",
			wantMessage: "resource not found",
		},
		{
			name:        "method not allowed",
			request:     httptest.NewRequest(stdhttp.MethodGet, "/v1/teams", nil),
			wantStatus:  stdhttp.StatusMethodNotAllowed,
			wantCode:    "method_not_allowed",
			wantMessage: "method not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, tt.request)

			assertErrorResponse(
				t,
				recorder,
				tt.wantStatus,
				tt.wantCode,
				tt.wantMessage,
			)
		})
	}
}

func newTestHandler(service teamService) stdhttp.Handler {
	return requestid.Middleware(NewHandler(service))
}

func assertErrorResponse(
	t *testing.T,
	recorder *httptest.ResponseRecorder,
	wantStatus int,
	wantCode string,
	wantMessage string,
) {
	t.Helper()

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != wantStatus {
		t.Fatalf("status = %d, want %d", response.StatusCode, wantStatus)
	}

	requestID := response.Header.Get(requestid.Header)
	if requestID == "" {
		t.Fatalf("%s response header = empty, want generated ID", requestid.Header)
	}

	var got errorResponse
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if got.Code != wantCode {
		t.Fatalf("error code = %q, want %q", got.Code, wantCode)
	}
	if got.Message != wantMessage {
		t.Fatalf("error message = %q, want %q", got.Message, wantMessage)
	}
	if got.RequestID != requestID {
		t.Fatalf("error request ID = %q, want response header %q", got.RequestID, requestID)
	}
}

type recordingTeamService struct {
	createInput *catalog.CreateTeamInput
	team        catalog.Team
	err         error
}

func (s *recordingTeamService) CreateTeam(
	_ context.Context,
	input catalog.CreateTeamInput,
) (catalog.Team, error) {
	s.createInput = &input
	return s.team, s.err
}
