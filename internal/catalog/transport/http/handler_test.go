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

func TestHandlerGetTeam(t *testing.T) {
	createdAt := time.Date(2026, time.August, 10, 12, 0, 0, 0, time.FixedZone("CST", 8*60*60))
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

	request := httptest.NewRequest(stdhttp.MethodGet, "/v1/teams/42", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != stdhttp.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, stdhttp.StatusOK)
	}
	if service.getID == nil {
		t.Fatal("GetTeamByID() was not called")
	}
	if *service.getID != 42 {
		t.Fatalf("GetTeamByID() ID = %d, want 42", *service.getID)
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
	if !got.UpdatedAt.Equal(createdAt.Add(time.Hour).UTC()) {
		t.Fatalf("response updated_at = %s, want %s", got.UpdatedAt, createdAt.Add(time.Hour).UTC())
	}
}

func TestHandlerGetTeamRejectsInvalidID(t *testing.T) {
	service := &recordingTeamService{}
	handler := newTestHandler(service)

	for _, path := range []string{
		"/v1/teams/0",
		"/v1/teams/-1",
		"/v1/teams/not-a-number",
	} {
		t.Run(path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(stdhttp.MethodGet, path, nil)

			handler.ServeHTTP(recorder, request)

			if service.getID != nil {
				t.Fatalf("GetTeamByID() ID = %d, want no call", *service.getID)
			}
			assertErrorResponse(
				t,
				recorder,
				stdhttp.StatusBadRequest,
				"invalid_argument",
				"team ID must be a positive integer",
			)
		})
	}
}

func TestHandlerGetTeamMapsNotFound(t *testing.T) {
	service := &recordingTeamService{
		err: catalog.ErrNotFound,
	}
	handler := newTestHandler(service)

	request := httptest.NewRequest(stdhttp.MethodGet, "/v1/teams/42", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if service.getID == nil {
		t.Fatal("GetTeamByID() was not called")
	}
	if *service.getID != 42 {
		t.Fatalf("GetTeamByID() ID = %d, want 42", *service.getID)
	}
	assertErrorResponse(
		t,
		recorder,
		stdhttp.StatusNotFound,
		"not_found",
		"resource not found",
	)
}

func TestHandlerListTeams(t *testing.T) {
	firstCreatedAt := time.Date(
		2026,
		time.August,
		11,
		21,
		0,
		0,
		0,
		time.FixedZone("CST", 8*60*60),
	)
	secondCreatedAt := firstCreatedAt.Add(-time.Minute)

	service := &recordingTeamService{
		page: catalog.TeamPage{
			Teams: []catalog.Team{
				{
					ID:        42,
					Slug:      "platform",
					Name:      "Platform",
					CreatedAt: firstCreatedAt,
					UpdatedAt: firstCreatedAt,
				},
				{
					ID:        41,
					Slug:      "observability",
					Name:      "Observability",
					CreatedAt: secondCreatedAt,
					UpdatedAt: secondCreatedAt,
				},
			},
			Next: &catalog.TeamCursor{
				CreatedAt: secondCreatedAt,
				ID:        41,
			},
		},
	}
	handler := newTestHandler(service)

	request := httptest.NewRequest(stdhttp.MethodGet, "/v1/teams?limit=2", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != stdhttp.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, stdhttp.StatusOK)
	}

	if service.listInput == nil {
		t.Fatal("ListTeams() was not called")
	}
	if service.listInput.Limit != 2 {
		t.Fatalf("ListTeams() limit = %d, want 2", service.listInput.Limit)
	}
	if service.listInput.After != nil {
		t.Fatalf("ListTeams() cursor = %#v, want nil", service.listInput.After)
	}

	var got teamPageResponse
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if len(got.Items) != 2 {
		t.Fatalf("response item count = %d, want 2", len(got.Items))
	}
	if got.Items[0].ID != 42 || got.Items[0].Slug != "platform" {
		t.Fatalf("first response item = %#v, want platform team", got.Items[0])
	}
	if got.Items[1].ID != 41 || got.Items[1].Slug != "observability" {
		t.Fatalf("second response item = %#v, want observability team", got.Items[1])
	}
	if got.Items[0].CreatedAt.Location() != time.UTC {
		t.Fatalf("first response created_at location = %s, want UTC", got.Items[0].CreatedAt.Location())
	}

	next, err := decodeTeamCursor(got.NextCursor)
	if err != nil {
		t.Fatalf("decode response next_cursor: %v", err)
	}
	if next.ID != 41 {
		t.Fatalf("response next cursor ID = %d, want 41", next.ID)
	}
	if !next.CreatedAt.Equal(secondCreatedAt) {
		t.Fatalf(
			"response next cursor created_at = %s, want %s",
			next.CreatedAt,
			secondCreatedAt,
		)
	}
}

func TestHandlerListTeamsPassesDefaultAndCursorInput(t *testing.T) {
	after := catalog.TeamCursor{
		CreatedAt: time.Date(
			2026,
			time.August,
			11,
			12,
			0,
			0,
			123,
			time.FixedZone("CST", 8*60*60),
		),
		ID: 41,
	}
	cursor, err := encodeTeamCursor(after)
	if err != nil {
		t.Fatalf("encodeTeamCursor() error = %v", err)
	}

	tests := []struct {
		name      string
		path      string
		wantLimit int32
		wantAfter *catalog.TeamCursor
	}{
		{
			name:      "default limit",
			path:      "/v1/teams",
			wantLimit: 0,
		},
		{
			name:      "cursor",
			path:      "/v1/teams?cursor=" + cursor,
			wantLimit: 0,
			wantAfter: &after,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &recordingTeamService{}
			handler := newTestHandler(service)

			request := httptest.NewRequest(stdhttp.MethodGet, tt.path, nil)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if recorder.Code != stdhttp.StatusOK {
				t.Fatalf("status = %d, want %d", recorder.Code, stdhttp.StatusOK)
			}
			if service.listInput == nil {
				t.Fatal("ListTeams() was not called")
			}
			if service.listInput.Limit != tt.wantLimit {
				t.Fatalf(
					"ListTeams() limit = %d, want %d",
					service.listInput.Limit,
					tt.wantLimit,
				)
			}

			if tt.wantAfter == nil {
				if service.listInput.After != nil {
					t.Fatalf(
						"ListTeams cursor = %#v, want nil",
						service.listInput.After,
					)
				}
				return
			}

			if service.listInput.After == nil {
				t.Fatal("ListTeams() cursor = nil, want value")
			}
			if service.listInput.After.ID != tt.wantAfter.ID {
				t.Fatalf(
					"ListTeams() cursor ID = %d, want %d",
					service.listInput.After.ID,
					tt.wantAfter.ID,
				)
			}
			if !service.listInput.After.CreatedAt.Equal(tt.wantAfter.CreatedAt) {
				t.Fatalf(
					"ListTeams cursor created_at = %s, want %s",
					service.listInput.After.CreatedAt,
					tt.wantAfter.CreatedAt,
				)
			}
			if service.listInput.After.CreatedAt.Location() != time.UTC {
				t.Fatalf(
					"ListTeams cursor location = %s, want UTC",
					service.listInput.After.CreatedAt.Location(),
				)
			}
		})
	}
}

func TestHandlerListTeamsRejectsInvalidQuery(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		wantMessage string
	}{
		{
			name:        "non-integer limit",
			path:        "/v1/teams?limit=two",
			wantMessage: "limit must be an integer",
		},
		{
			name:        "invalid cursor",
			path:        "/v1/teams?cursor=not-a-valid-cursor!",
			wantMessage: "cursor is invalid",
		},
		{
			name:        "zero limit",
			path:        "/v1/teams?limit=0",
			wantMessage: "limit must be between 1 and 100",
		},
		{
			name:        "empty limit",
			path:        "/v1/teams?limit=",
			wantMessage: "limit must be an integer",
		},
		{
			name:        "empty cursor",
			path:        "/v1/teams?cursor=",
			wantMessage: "cursor is invalid",
		},
		{
			name:        "duplicate limit",
			path:        "/v1/teams?limit=1&limit=2",
			wantMessage: "limit must be provided once",
		},
		{
			name:        "duplicate cursor",
			path:        "/v1/teams?cursor=first&cursor=second",
			wantMessage: "cursor must be provided once",
		},
		{
			name:        "limit above maximum",
			path:        "/v1/teams?limit=2147483648",
			wantMessage: "limit must be between 1 and 100",
		},
		{
			name:        "limit overflows int64",
			path:        "/v1/teams?limit=9223372036854775808",
			wantMessage: "limit must be between 1 and 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &recordingTeamService{}
			handler := newTestHandler(service)

			request := httptest.NewRequest(stdhttp.MethodGet, tt.path, nil)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if service.listInput != nil {
				t.Fatalf("ListTeams() input = %#v, want no call", *service.listInput)
			}
			assertErrorResponse(
				t,
				recorder,
				stdhttp.StatusBadRequest,
				"invalid_argument",
				tt.wantMessage,
			)
		})
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
			request:     httptest.NewRequest(stdhttp.MethodPut, "/v1/teams", nil),
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
	getID       *int64
	listInput   *catalog.ListTeamsInput
	team        catalog.Team
	page        catalog.TeamPage
	err         error
}

func (s *recordingTeamService) CreateTeam(
	_ context.Context,
	input catalog.CreateTeamInput,
) (catalog.Team, error) {
	s.createInput = &input
	return s.team, s.err
}

func (s *recordingTeamService) GetTeamByID(
	_ context.Context,
	id int64,
) (catalog.Team, error) {
	s.getID = &id
	return s.team, s.err
}

func (s *recordingTeamService) ListTeams(
	_ context.Context,
	input catalog.ListTeamsInput,
) (catalog.TeamPage, error) {
	s.listInput = &input
	return s.page, s.err
}
