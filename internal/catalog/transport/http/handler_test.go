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

	request := newJSONRequest(
		stdhttp.MethodPost,
		"/v1/teams",
		`{"slug":"platform","name":"Platform"}`,
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

func TestHandlerUpdateTeam(t *testing.T) {
	createdAt := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)

	service := &recordingTeamService{
		team: catalog.Team{
			ID:        42,
			Slug:      "platform-engineering",
			Name:      "Platform Engineering",
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		},
	}
	handler := newTestHandler(service)

	request := newJSONRequest(
		stdhttp.MethodPatch,
		"/v1/teams/42",
		`{"slug":"platform-engineering","name":"Platform Engineering"}`,
	)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != stdhttp.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, stdhttp.StatusOK)
	}
	if service.updateID == nil {
		t.Fatal("UpdateTeam() was not called")
	}
	if *service.updateID != 42 {
		t.Fatalf("UpdateTeam() ID = %d, want 42", *service.updateID)
	}
	if service.updateInput == nil {
		t.Fatal("UpdateTeam() input = nil")
	}
	if service.updateInput.Slug == nil {
		t.Fatal("UpdateTeam() slug = nil")
	}
	if *service.updateInput.Slug != "platform-engineering" {
		t.Fatalf(
			"UpdateTeam() slug = %q, want %q",
			*service.updateInput.Slug,
			"platform-engineering",
		)
	}
	if service.updateInput.Name == nil {
		t.Fatal("UpdateTeam() name = nil")
	}
	if *service.updateInput.Name != "Platform Engineering" {
		t.Fatalf(
			"UpdateTeam() name = %q, want %q",
			*service.updateInput.Name,
			"Platform Engineering",
		)
	}

	var got teamResponse
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if got.ID != 42 {
		t.Fatalf("response ID = %d, want 42", got.ID)
	}
	if got.Slug != "platform-engineering" {
		t.Fatalf("response slug = %q, want %q", got.Slug, "platform-engineering")
	}
	if got.Name != "Platform Engineering" {
		t.Fatalf("response name = %q, want %q", got.Name, "Platform Engineering")
	}
	if !got.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("response updated_at = %s, want %s", got.UpdatedAt, updatedAt)
	}
}

func TestHandlerUpdateTeamPassesPartialInput(t *testing.T) {
	service := &recordingTeamService{}
	handler := newTestHandler(service)

	request := newJSONRequest(
		stdhttp.MethodPatch,
		"/v1/teams/42",
		`{"name":"Platform Engineering"}`,
	)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != stdhttp.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, stdhttp.StatusOK)
	}
	if service.updateInput == nil {
		t.Fatal("UpdateTeam() was not called")
	}
	if service.updateInput.Slug != nil {
		t.Fatalf("UpdateTeam slug = %q, want nil", *service.updateInput.Slug)
	}
	if service.updateInput.Name == nil {
		t.Fatal("UpdateTeam() name = nil")
	}
	if *service.updateInput.Name != "Platform Engineering" {
		t.Fatalf(
			"UpdateTeam() name = %q, want %q",
			*service.updateInput.Name,
			"Platform Engineering",
		)
	}
}

func TestHandlerUpdateTeamRejectsInvalidRequestBody(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "null field",
			body: `{"name":null}`,
		},
		{
			name: "wrong field type",
			body: `{"name":42}`,
		},
		{
			name: "unknown field",
			body: `{"description":"platform team"}`,
		},
		{
			name: "multiple JSON values",
			body: `{"name":"Platform"} {"slug":"platform"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &recordingTeamService{}
			handler := newTestHandler(service)

			request := newJSONRequest(
				stdhttp.MethodPatch,
				"/v1/teams/42",
				tt.body,
			)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if service.updateInput != nil {
				t.Fatalf(
					"UpdateTeam() input = %#v, want no call",
					*service.updateInput,
				)
			}
			assertErrorResponse(
				t,
				recorder,
				stdhttp.StatusBadRequest,
				"invalid_argument",
				"invalid request body",
			)
		})
	}
}

func TestHandlerUpdateTeamMapsCatalogErrors(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		err         error
		wantStatus  int
		wantCode    string
		wantMessage string
	}{
		{
			name:        "empty update",
			body:        `{}`,
			err:         &catalog.InvalidArgumentError{Message: "at least one field must be provided"},
			wantStatus:  stdhttp.StatusBadRequest,
			wantCode:    "invalid_argument",
			wantMessage: "at least one field must be provided",
		},
		{
			name:        "not found",
			body:        `{"name":"Platform Engineering"}`,
			err:         catalog.ErrNotFound,
			wantStatus:  stdhttp.StatusNotFound,
			wantCode:    "not_found",
			wantMessage: "resource not found",
		},
		{
			name:        "conflict",
			body:        `{"slug":"platform-engineering"}`,
			err:         catalog.ErrConflict,
			wantStatus:  stdhttp.StatusConflict,
			wantCode:    "conflict",
			wantMessage: "resource conflict",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &recordingTeamService{
				err: tt.err,
			}
			handler := newTestHandler(service)

			request := newJSONRequest(
				stdhttp.MethodPatch,
				"/v1/teams/42",
				tt.body,
			)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if service.updateID == nil {
				t.Fatal("UpdateTeam() was not called")
			}
			if *service.updateID != 42 {
				t.Fatalf("UpdateTeam() ID = %d, want 42", *service.updateID)
			}
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

func TestHandlerDeleteTeam(t *testing.T) {
	service := &recordingTeamService{}
	handler := newTestHandler(service)

	request := httptest.NewRequest(
		stdhttp.MethodDelete,
		"/v1/teams/42",
		nil,
	)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != stdhttp.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, stdhttp.StatusNoContent)
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("response body = %q, want empty", recorder.Body.String())
	}
	if service.deleteID == nil {
		t.Fatal("DeleteTeam() was not called")
	}
	if *service.deleteID != 42 {
		t.Fatalf("DeleteTeam() ID = %d, want 42", *service.deleteID)
	}
}

func TestHandlerDeleteTeamMapsCatalogErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantStatus  int
		wantCode    string
		wantMessage string
	}{
		{
			name:        "not found",
			err:         catalog.ErrNotFound,
			wantStatus:  stdhttp.StatusNotFound,
			wantCode:    "not_found",
			wantMessage: "resource not found",
		},
		{
			name:        "conflict",
			err:         catalog.ErrConflict,
			wantStatus:  stdhttp.StatusConflict,
			wantCode:    "conflict",
			wantMessage: "resource conflict",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &recordingTeamService{
				err: tt.err,
			}
			handler := newTestHandler(service)

			request := httptest.NewRequest(
				stdhttp.MethodDelete,
				"/v1/teams/42",
				nil,
			)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if service.deleteID == nil {
				t.Fatal("DeleteTeam() was not called")
			}
			if *service.deleteID != 42 {
				t.Fatalf("DeleteTeam() ID = %d, want 42", *service.deleteID)
			}
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

func TestHandlerCreateTeamRejectsInvalidRequestBody(t *testing.T) {
	service := &recordingTeamService{}
	handler := newTestHandler(service)

	request := newJSONRequest(
		stdhttp.MethodPost,
		"/v1/teams",
		`{"slug":"platform","name":"Platform","unexpected":"true"}`,
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

func TestHandlerCreateTeamRejectsUnsupportedMediaType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
	}{
		{
			name: "missing",
		},
		{
			name:        "plain text",
			contentType: "text/plain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &recordingTeamService{}
			handler := newTestHandler(service)

			request := httptest.NewRequest(
				stdhttp.MethodPost,
				"/v1/teams",
				strings.NewReader(`{"slug":"platform","name":"Platform"}`),
			)
			if tt.contentType != "" {
				request.Header.Set("Content-Type", tt.contentType)
			}

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)

			if service.createInput != nil {
				t.Fatalf(
					"CreateTeam() input = %#v, want no call",
					*service.createInput,
				)
			}
			assertErrorResponse(
				t,
				recorder,
				stdhttp.StatusUnsupportedMediaType,
				"unsupported_media_type",
				"content type must be application/json",
			)
		})
	}
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

			request := newJSONRequest(
				stdhttp.MethodPost,
				"/v1/teams",
				`{"slug":"platform","name":"Platform"}`,
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

func TestHandlerCreateService(t *testing.T) {
	createdAt := time.Date(
		2026,
		time.August,
		21,
		12,
		0,
		0,
		0,
		time.FixedZone("CST", 8*60*60),
	)
	description := "Owns the internal platform"

	services := &recordingServiceService{
		detail: catalog.ServiceDetail{
			Service: catalog.Service{
				ID:          42,
				TeamID:      7,
				Slug:        "catalog",
				Name:        "Service Catalog",
				Description: &description,
				CreatedAt:   createdAt,
				UpdatedAt:   createdAt.Add(time.Hour),
			},
			Environments: []catalog.Environment{
				{
					ID:        101,
					ServiceID: 42,
					Slug:      "production",
					Name:      "Production",
					CreatedAt: createdAt,
					UpdatedAt: createdAt,
				},
			},
		},
	}
	handler := newTestHandlerWithServices(&recordingTeamService{}, services)

	request := newJSONRequest(
		stdhttp.MethodPost,
		"/v1/services",
		`{"team_id":7,"slug":"catalog","name":"Service Catalog","description":"Owns the internal platform","environments":[{"slug":"production","name":"Production"}]}`,
	)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != stdhttp.StatusCreated {
		t.Fatalf("status = %d, want %d", response.StatusCode, stdhttp.StatusCreated)
	}
	if services.createInput == nil {
		t.Fatal("CreateService() was not called")
	}

	gotInput := *services.createInput
	if gotInput.TeamID != 7 {
		t.Fatalf("CreateService() team ID = %d, want 7", gotInput.TeamID)
	}
	if gotInput.Slug != "catalog" {
		t.Fatalf("CreateService() slug = %q, want %q", gotInput.Slug, "catalog")
	}
	if gotInput.Name != "Service Catalog" {
		t.Fatalf("CreateService() name = %q, want %q", gotInput.Name, "Service Catalog")
	}
	if gotInput.Description == nil || *gotInput.Description != description {
		t.Fatalf("CreateService() description = %#v, want %q", gotInput.Description, description)
	}
	if len(gotInput.Environments) != 1 {
		t.Fatalf("CreateService() environments = %#v, want one environment", gotInput.Environments)
	}
	if gotInput.Environments[0] != (catalog.CreateEnvironmentInput{
		Slug: "production",
		Name: "Production",
	}) {
		t.Fatalf(
			"CreateService() environment = %#v, want production",
			gotInput.Environments[0],
		)
	}

	var got serviceDetailResponse
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if got.ID != 42 || got.TeamID != 7 {
		t.Fatalf("response IDs = (%d, %d), want (42, 7)", got.ID, got.TeamID)
	}
	if got.Slug != "catalog" || got.Name != "Service Catalog" {
		t.Fatalf("response = %#v, want catalog service", got)
	}
	if got.Description == nil || *got.Description != description {
		t.Fatalf("response description = %#v, want %q", got.Description, description)
	}
	if !got.CreatedAt.Equal(createdAt.UTC()) {
		t.Fatalf("response created_at = %s, want %s", got.CreatedAt, createdAt.UTC())
	}
	if got.UpdatedAt.Location() != time.UTC {
		t.Fatalf("response updated_at location = %s, want UTC", got.UpdatedAt.Location())
	}
	if len(got.Environments) != 1 {
		t.Fatalf("response environments = %#v, want one environment", got.Environments)
	}
	if got.Environments[0].ID != 101 || got.Environments[0].ServiceID != 42 {
		t.Fatalf("response environment = %#v, want IDs (101, 42)", got.Environments[0])
	}
}

func TestHandlerCreateServiceRejectsInvalidRequestBody(t *testing.T) {
	services := &recordingServiceService{}
	handler := newTestHandlerWithServices(&recordingTeamService{}, services)

	request := newJSONRequest(
		stdhttp.MethodPost,
		"/v1/services",
		`{"team_id":7,"slug":"catalog","name":"Service Catalog","unexpected":true}`,
	)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if services.createInput != nil {
		t.Fatalf("CreateService() input = %#v, want no call", *services.createInput)
	}
	assertErrorResponse(
		t,
		recorder,
		stdhttp.StatusBadRequest,
		"invalid_argument",
		"invalid request body",
	)
}

func TestHandlerCreateServiceRejectsUnsupportedMediaType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
	}{
		{
			name: "missing",
		},
		{
			name:        "plain text",
			contentType: "text/plain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			services := &recordingServiceService{}
			handler := newTestHandlerWithServices(
				&recordingTeamService{},
				services,
			)

			request := httptest.NewRequest(
				stdhttp.MethodPost,
				"/v1/services",
				strings.NewReader(
					`{"team_id":7,"slug":"catalog","name":"Service Catalog"}`,
				),
			)
			if tt.contentType != "" {
				request.Header.Set("Content-Type", tt.contentType)
			}

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)

			if services.createInput != nil {
				t.Fatalf(
					"CreateService() input = %#v, want no call",
					*services.createInput,
				)
			}
			assertErrorResponse(
				t,
				recorder,
				stdhttp.StatusUnsupportedMediaType,
				"unsupported_media_type",
				"content type must be application/json",
			)
		})
	}
}

func TestHandlerCreateServiceMapsCatalogErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantStatus  int
		wantCode    string
		wantMessage string
	}{
		{
			name: "invalid argument",
			err: &catalog.InvalidArgumentError{
				Message: "team ID must be a positive integer",
			},
			wantStatus:  stdhttp.StatusBadRequest,
			wantCode:    "invalid_argument",
			wantMessage: "team ID must be a positive integer",
		},
		{
			name:        "team not found",
			err:         catalog.ErrNotFound,
			wantStatus:  stdhttp.StatusNotFound,
			wantCode:    "not_found",
			wantMessage: "resource not found",
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
			services := &recordingServiceService{
				err: tt.err,
			}
			handler := newTestHandlerWithServices(
				&recordingTeamService{},
				services,
			)

			request := newJSONRequest(
				stdhttp.MethodPost,
				"/v1/services",
				`{"team_id":7,"slug":"catalog","name":"Service Catalog"}`,
			)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if services.createInput == nil {
				t.Fatal("CreateService() was not called")
			}
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

func TestHandlerGetService(t *testing.T) {
	createdAt := time.Date(
		2026,
		time.August,
		21,
		13,
		0,
		0,
		0,
		time.FixedZone("CST", 8*60*60),
	)
	description := "Owns the internal platform"

	services := &recordingServiceService{
		detail: catalog.ServiceDetail{
			Service: catalog.Service{
				ID:          42,
				TeamID:      7,
				Slug:        "catalog",
				Name:        "Service Catalog",
				Description: &description,
				CreatedAt:   createdAt,
				UpdatedAt:   createdAt.Add(time.Hour),
			},
			Environments: []catalog.Environment{
				{
					ID:        101,
					ServiceID: 42,
					Slug:      "production",
					Name:      "Production",
					CreatedAt: createdAt,
					UpdatedAt: createdAt,
				},
			},
		},
	}
	handler := newTestHandlerWithServices(&recordingTeamService{}, services)

	request := httptest.NewRequest(
		stdhttp.MethodGet,
		"/v1/services/42",
		nil,
	)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != stdhttp.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, stdhttp.StatusOK)
	}
	if services.getID == nil {
		t.Fatal("GetServiceByID was not called")
	}
	if *services.getID != 42 {
		t.Fatalf("GetServiceByID() ID = %d, want 42", *services.getID)
	}

	var got serviceDetailResponse
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if got.ID != 42 || got.TeamID != 7 {
		t.Fatalf("response IDs = (%d, %d), want (42, 7)", got.ID, got.TeamID)
	}
	if got.Description == nil || *got.Description != description {
		t.Fatalf("response description = %#v, want %q", got.Description, description)
	}
	if !got.CreatedAt.Equal(createdAt.UTC()) {
		t.Fatalf("response created_at = %s, want %s", got.CreatedAt, createdAt.UTC())
	}
	if len(got.Environments) != 1 {
		t.Fatalf("response environments = %#v, want one environment", got.Environments)
	}
	if got.Environments[0].Slug != "production" {
		t.Fatalf(
			"response environment slug = %q, want %q",
			got.Environments[0].Slug,
			"production",
		)
	}
}

func TestHandlerGetServiceRejectsInvalidID(t *testing.T) {
	services := &recordingServiceService{}
	handler := newTestHandlerWithServices(&recordingTeamService{}, services)

	for _, path := range []string{
		"/v1/services/0",
		"/v1/services/-1",
		"/v1/services/not-a-number",
	} {
		t.Run(path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(stdhttp.MethodGet, path, nil)

			handler.ServeHTTP(recorder, request)

			if services.getID != nil {
				t.Fatalf(
					"GetServiceByID() ID = %d, want no call",
					*services.getID,
				)
			}
			assertErrorResponse(
				t,
				recorder,
				stdhttp.StatusBadRequest,
				"invalid_argument",
				"service ID must be a positive integer",
			)
		})
	}
}

func TestHandlerGetServiceMapsCatalogErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantStatus  int
		wantCode    string
		wantMessage string
	}{
		{
			name:        "not found",
			err:         catalog.ErrNotFound,
			wantStatus:  stdhttp.StatusNotFound,
			wantCode:    "not_found",
			wantMessage: "resource not found",
		},
		{
			name:        "unknown error",
			err:         errors.New("database connection failed"),
			wantStatus:  stdhttp.StatusInternalServerError,
			wantCode:    "internal",
			wantMessage: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			services := &recordingServiceService{
				err: tt.err,
			}
			handler := newTestHandlerWithServices(
				&recordingTeamService{},
				services,
			)

			request := httptest.NewRequest(
				stdhttp.MethodGet,
				"/v1/services/42",
				nil,
			)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if services.getID == nil {
				t.Fatal("GetServiceByID() was not called")
			}
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

func TestHandlerListServices(t *testing.T) {
	firstCreatedAt := time.Date(
		2026,
		time.August,
		21,
		14,
		0,
		0,
		0,
		time.FixedZone("CST", 8*60*60),
	)
	secondCreatedAt := firstCreatedAt.Add(-time.Minute)
	description := "Owns the internal platform"

	services := &recordingServiceService{
		page: catalog.ServicePage{
			Services: []catalog.Service{
				{
					ID:          42,
					TeamID:      7,
					Slug:        "catalog",
					Name:        "Service Catalog",
					Description: &description,
					CreatedAt:   firstCreatedAt,
					UpdatedAt:   firstCreatedAt,
				},
				{
					ID:        41,
					TeamID:    7,
					Slug:      "observability",
					Name:      "Observability",
					CreatedAt: secondCreatedAt,
					UpdatedAt: secondCreatedAt,
				},
			},
			Next: &catalog.ServiceCursor{
				CreatedAt: secondCreatedAt,
				ID:        41,
			},
		},
	}
	handler := newTestHandlerWithServices(&recordingTeamService{}, services)

	request := httptest.NewRequest(
		stdhttp.MethodGet,
		"/v1/services?team_id=7&limit=2",
		nil,
	)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != stdhttp.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, stdhttp.StatusOK)
	}
	if services.listInput == nil {
		t.Fatal("ListServices() was not called")
	}
	if services.listInput.TeamID == nil {
		t.Fatal("ListServices() team ID = nil, want 7")
	}
	if *services.listInput.TeamID != 7 {
		t.Fatalf("ListServices() team ID = %d, want 7", *services.listInput.TeamID)
	}
	if services.listInput.Limit != 2 {
		t.Fatalf("ListServices() limit = %d, want 2", services.listInput.Limit)
	}
	if services.listInput.After != nil {
		t.Fatalf("ListServices() cursor = %#v, want nil", services.listInput.After)
	}

	var got servicePageResponse
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if len(got.Items) != 2 {
		t.Fatalf("response item count = %d, want 2", len(got.Items))
	}
	if got.Items[0].ID != 42 || got.Items[0].Slug != "catalog" {
		t.Fatalf("first response item = %#v, want catalog service", got.Items[0])
	}
	if got.Items[1].ID != 41 || got.Items[1].Slug != "observability" {
		t.Fatalf(
			"second response item = %#v, want observability service",
			got.Items[1],
		)
	}
	if got.Items[0].Description == nil || *got.Items[0].Description != description {
		t.Fatalf(
			"first response description = %#v, want %q",
			got.Items[0].Description,
			description,
		)
	}
	if got.Items[0].CreatedAt.Location() != time.UTC {
		t.Fatalf(
			"first response created_at location = %s, want UTC",
			got.Items[0].CreatedAt.Location(),
		)
	}

	next, err := decodeServiceCursor(got.NextCursor)
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

func TestHandlerListServicesPassesDefaultCursorAndTeamFilter(t *testing.T) {
	after := catalog.ServiceCursor{
		CreatedAt: time.Date(
			2026,
			time.August,
			22,
			12,
			0,
			0,
			123,
			time.FixedZone("CST", 8*60*60),
		),
		ID: 41,
	}
	cursor, err := encodeServiceCursor(after)
	if err != nil {
		t.Fatalf("encodeServiceCursor() error = %v", err)
	}

	teamID := int64(7)
	tests := []struct {
		name       string
		path       string
		wantTeamID *int64
		wantLimit  int32
		wantAfter  *catalog.ServiceCursor
	}{
		{
			name:      "default limit without filter",
			path:      "/v1/services",
			wantLimit: 0,
		},
		{
			name:       "team filter",
			path:       "/v1/services?team_id=7",
			wantTeamID: &teamID,
			wantLimit:  0,
		},
		{
			name:       "team filter and cursor",
			path:       "/v1/services?team_id=7&cursor=" + cursor,
			wantTeamID: &teamID,
			wantLimit:  0,
			wantAfter:  &after,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			services := &recordingServiceService{}
			handler := newTestHandlerWithServices(
				&recordingTeamService{},
				services,
			)

			request := httptest.NewRequest(stdhttp.MethodGet, tt.path, nil)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if recorder.Code != stdhttp.StatusOK {
				t.Fatalf("status = %d, want %d", recorder.Code, stdhttp.StatusOK)
			}
			if services.listInput == nil {
				t.Fatal("ListServices() was not called")
			}
			if services.listInput.Limit != tt.wantLimit {
				t.Fatalf(
					"ListServices() limit = %d, want %d",
					services.listInput.Limit,
					tt.wantLimit,
				)
			}

			if tt.wantTeamID == nil {
				if services.listInput.TeamID != nil {
					t.Fatalf(
						"ListServices() team ID = %#v, want nil",
						services.listInput.TeamID,
					)
				}
			} else {
				if services.listInput.TeamID == nil {
					t.Fatal("ListServices() team ID = nil, want value")
				}
				if *services.listInput.TeamID != *tt.wantTeamID {
					t.Fatalf(
						"ListServices() team ID = %d, want %d",
						*services.listInput.TeamID,
						*tt.wantTeamID,
					)
				}
			}

			if tt.wantAfter == nil {
				if services.listInput.After != nil {
					t.Fatalf(
						"ListServices() cursor = %#v, want nil",
						services.listInput.After,
					)
				}
				return
			}

			if services.listInput.After == nil {
				t.Fatal("ListServices() cursor = nil, want value")
			}
			if services.listInput.After.ID != tt.wantAfter.ID {
				t.Fatalf(
					"ListServices() cursor ID = %d, want %d",
					services.listInput.After.ID,
					tt.wantAfter.ID,
				)
			}
			if !services.listInput.After.CreatedAt.Equal(tt.wantAfter.CreatedAt) {
				t.Fatalf(
					"ListServices() cursor created_at = %s, want %s",
					services.listInput.After.CreatedAt,
					tt.wantAfter.CreatedAt,
				)
			}
			if services.listInput.After.CreatedAt.Location() != time.UTC {
				t.Fatalf(
					"ListServices() cursor location = %s, want UTC",
					services.listInput.After.CreatedAt.Location(),
				)
			}
		})
	}
}

func TestHandlerListServicesRejectsInvalidQuery(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		wantMessage string
	}{
		{
			name:        "empty team ID",
			path:        "/v1/services?team_id=",
			wantMessage: "team ID must be a positive integer",
		},
		{
			name:        "non-positive team ID",
			path:        "/v1/services?team_id=0",
			wantMessage: "team ID must be a positive integer",
		},
		{
			name:        "non-integer team ID",
			path:        "/v1/services?team_id=seven",
			wantMessage: "team ID must be a positive integer",
		},
		{
			name:        "duplicate team ID",
			path:        "/v1/services?team_id=7&team_id=8",
			wantMessage: "team_id must be provided once",
		},
		{
			name:        "non-integer limit",
			path:        "/v1/services?limit=two",
			wantMessage: "limit must be an integer",
		},
		{
			name:        "zero limit",
			path:        "/v1/services?limit=0",
			wantMessage: "limit must be between 1 and 100",
		},
		{
			name:        "empty limit",
			path:        "/v1/services?limit=",
			wantMessage: "limit must be an integer",
		},
		{
			name:        "duplicate limit",
			path:        "/v1/services?limit=1&limit=2",
			wantMessage: "limit must be provided once",
		},
		{
			name:        "limit above maximum",
			path:        "/v1/services?limit=2147483648",
			wantMessage: "limit must be between 1 and 100",
		},
		{
			name:        "limit overflows int64",
			path:        "/v1/services?limit=9223372036854775808",
			wantMessage: "limit must be between 1 and 100",
		},
		{
			name:        "invalid cursor",
			path:        "/v1/services?cursor=not-a-valid-cursor!",
			wantMessage: "cursor is invalid",
		},
		{
			name:        "empty cursor",
			path:        "/v1/services?cursor=",
			wantMessage: "cursor is invalid",
		},
		{
			name:        "duplicate cursor",
			path:        "/v1/services?cursor=first&cursor=second",
			wantMessage: "cursor must be provided once",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			services := &recordingServiceService{}
			handler := newTestHandlerWithServices(
				&recordingTeamService{},
				services,
			)

			request := httptest.NewRequest(stdhttp.MethodGet, tt.path, nil)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if services.listInput != nil {
				t.Fatalf(
					"ListServices() input = %#v, want no call",
					*services.listInput,
				)
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

func TestHandlerUpdateService(t *testing.T) {
	createdAt := time.Date(2026, time.August, 24, 12, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	description := "Updated service description"

	service := &recordingServiceService{
		service: catalog.Service{
			ID:          42,
			TeamID:      7,
			Slug:        "catalog-api",
			Name:        "Catalog API",
			Description: &description,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		},
	}
	handler := newTestHandlerWithServices(&recordingTeamService{}, service)

	request := newJSONRequest(
		stdhttp.MethodPatch,
		"/v1/services/42",
		`{"slug":"catalog-api","name":"Catalog API","description":"Updated service description"}`,
	)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != stdhttp.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, stdhttp.StatusOK)
	}
	if service.updateID == nil {
		t.Fatal("UpdateService() was not called")
	}
	if *service.updateID != 42 {
		t.Fatalf("UpdateService() ID = %d, want 42", *service.updateID)
	}
	if service.updateInput == nil {
		t.Fatal("UpdateService() input = nil")
	}
	if service.updateInput.Slug == nil || *service.updateInput.Slug != "catalog-api" {
		t.Fatalf("UpdateService() Slug = %#v, want catalog-api", service.updateInput.Slug)
	}
	if service.updateInput.Name == nil || *service.updateInput.Name != "Catalog API" {
		t.Fatalf("UpdateService() Name = %#v, want Catalog API", service.updateInput.Name)
	}
	if !service.updateInput.DescriptionProvided {
		t.Fatal("UpdateService() DescriptionProvided = false, want true")
	}
	if service.updateInput.Description == nil ||
		*service.updateInput.Description != "Updated service description" {
		t.Fatalf(
			"UpdateService() Description = %#v, want Updated service description",
			service.updateInput.Description,
		)
	}

	var got serviceResponse
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if got.ID != 42 || got.TeamID != 7 {
		t.Fatalf("response IDs = (%d, %d), want (42, 7)", got.ID, got.TeamID)
	}
	if got.Slug != "catalog-api" || got.Name != "Catalog API" {
		t.Fatalf("response = %#v, want catalog-api / Catalog API", got)
	}
	if got.Description == nil || *got.Description != description {
		t.Fatalf("response Description = %#v, want %q", got.Description, description)
	}
	if !got.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("response updated_at = %s, want %s", got.UpdatedAt, updatedAt)
	}
}

func TestHandlerUpdateServicePassesExplicitNullDescription(t *testing.T) {
	service := &recordingServiceService{}
	handler := newTestHandlerWithServices(&recordingTeamService{}, service)

	request := newJSONRequest(
		stdhttp.MethodPatch,
		"/v1/services/42",
		`{"description":null}`,
	)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != stdhttp.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, stdhttp.StatusOK)
	}
	if service.updateInput == nil {
		t.Fatal("UpdateService was not called")
	}
	if service.updateInput.Slug != nil || service.updateInput.Name != nil {
		t.Fatalf("UpdateService() input %#v, want description only", *service.updateInput)
	}
	if !service.updateInput.DescriptionProvided {
		t.Fatal("UpdateService() DescriptionProvided = false, want true")
	}
	if service.updateInput.Description != nil {
		t.Fatalf("UpdateService() Description = %#v, want nil", service.updateInput.Description)
	}
}

func TestHandlerDeleteService(t *testing.T) {
	service := &recordingServiceService{}
	handler := newTestHandlerWithServices(&recordingTeamService{}, service)

	request := httptest.NewRequest(
		stdhttp.MethodDelete,
		"/v1/services/42",
		nil,
	)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != stdhttp.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, stdhttp.StatusNoContent)
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("response body = %q, want empty", recorder.Body.String())
	}
	if service.deleteID == nil {
		t.Fatal("DeleteService() was not called")
	}
	if *service.deleteID != 42 {
		t.Fatalf("DeleteService() ID = %d, want 42", *service.deleteID)
	}
}

func TestHandlerUpdateServiceRejectsInvalidRequestBody(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "null slug",
			body: `{"slug":null}`,
		},
		{
			name: "wrong description type",
			body: `{"description":42}`,
		},
		{
			name: "unknown field",
			body: `{"environment":"production"}`,
		},
		{
			name: "multiple JSON values",
			body: `{"name":"Catalog API"} {"slug":"catalog-api"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &recordingServiceService{}
			handler := newTestHandlerWithServices(&recordingTeamService{}, service)

			request := newJSONRequest(
				stdhttp.MethodPatch,
				"/v1/services/42",
				tt.body,
			)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if service.updateInput != nil {
				t.Fatalf(
					"UpdateService() input = %#v, want no call",
					*service.updateInput,
				)
			}
			assertErrorResponse(
				t,
				recorder,
				stdhttp.StatusBadRequest,
				"invalid_argument",
				"invalid request body",
			)
		})
	}
}

func TestHandlerUpdateServiceRejectsUnsupportedMediaType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
	}{
		{
			name: "missing",
		},
		{
			name:        "plain text",
			contentType: "text/plain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &recordingServiceService{}
			handler := newTestHandlerWithServices(&recordingTeamService{}, service)

			request := httptest.NewRequest(
				stdhttp.MethodPatch,
				"/v1/services/42",
				strings.NewReader(`{"name":"Catalog API"}`),
			)
			if tt.contentType != "" {
				request.Header.Set("Content-Type", tt.contentType)
			}
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if service.updateInput != nil {
				t.Fatalf(
					"UpdateService() input = %#v, want no call",
					*service.updateInput,
				)
			}
			assertErrorResponse(
				t,
				recorder,
				stdhttp.StatusUnsupportedMediaType,
				"unsupported_media_type",
				"content type must be application/json",
			)
		})
	}
}

func TestHandlerUpdateServiceMapsCatalogErrors(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		err         error
		wantStatus  int
		wantCode    string
		wantMessage string
	}{
		{
			name:        "empty update",
			body:        `{}`,
			err:         &catalog.InvalidArgumentError{Message: "at least one field must be provided"},
			wantStatus:  stdhttp.StatusBadRequest,
			wantCode:    "invalid_argument",
			wantMessage: "at least one field must be provided",
		},
		{
			name:        "not found",
			body:        `{"name":"Catalog API"}`,
			err:         catalog.ErrNotFound,
			wantStatus:  stdhttp.StatusNotFound,
			wantCode:    "not_found",
			wantMessage: "resource not found",
		},
		{
			name:        "conflict",
			body:        `{"slug":"catalog-api"}`,
			err:         catalog.ErrConflict,
			wantStatus:  stdhttp.StatusConflict,
			wantCode:    "conflict",
			wantMessage: "resource conflict",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &recordingServiceService{
				err: tt.err,
			}
			handler := newTestHandlerWithServices(&recordingTeamService{}, service)

			request := newJSONRequest(
				stdhttp.MethodPatch,
				"/v1/services/42",
				tt.body,
			)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if service.updateID == nil {
				t.Fatal("UpdateService() was not called")
			}
			if *service.updateID != 42 {
				t.Fatalf("UpdateService() ID = %d, want 42", *service.updateID)
			}
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

func TestHandlerDeleteServiceMapsCatalogErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantStatus  int
		wantCode    string
		wantMessage string
	}{
		{
			name:        "not found",
			err:         catalog.ErrNotFound,
			wantStatus:  stdhttp.StatusNotFound,
			wantCode:    "not_found",
			wantMessage: "resource not found",
		},
		{
			name:        "conflict",
			err:         catalog.ErrConflict,
			wantStatus:  stdhttp.StatusConflict,
			wantCode:    "conflict",
			wantMessage: "resource conflict",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &recordingServiceService{
				err: tt.err,
			}
			handler := newTestHandlerWithServices(&recordingTeamService{}, service)

			request := httptest.NewRequest(
				stdhttp.MethodDelete,
				"/v1/services/42",
				nil,
			)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if service.deleteID == nil {
				t.Fatal("DeleteService() was not called")
			}
			if *service.deleteID != 42 {
				t.Fatalf("DeleteService() ID = %d, want 42", *service.deleteID)
			}
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

func TestHandlerServiceMutationsRejectInvalidID(t *testing.T) {
	tests := []struct {
		name   string
		method string
	}{
		{
			name:   "patch",
			method: stdhttp.MethodPatch,
		},
		{
			name:   "delete",
			method: stdhttp.MethodDelete,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &recordingServiceService{}
			handler := newTestHandlerWithServices(&recordingTeamService{}, service)

			var request *stdhttp.Request
			if tt.method == stdhttp.MethodPatch {
				request = newJSONRequest(
					tt.method,
					"/v1/services/0",
					`{"name":"Catalog API"}`,
				)
			} else {
				request = httptest.NewRequest(
					tt.method,
					"/v1/services/0",
					nil,
				)
			}
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if service.updateID != nil {
				t.Fatalf("UpdateService() ID = %d, want no call", *service.updateID)
			}
			if service.deleteID != nil {
				t.Fatalf("DeleteService() ID = %d, want no call", *service.deleteID)
			}
			assertErrorResponse(
				t,
				recorder,
				stdhttp.StatusBadRequest,
				"invalid_argument",
				"service ID must be a positive integer",
			)
		})
	}
}

func newJSONRequest(
	method string,
	target string,
	body string,
) *stdhttp.Request {
	request := httptest.NewRequest(
		method,
		target,
		strings.NewReader(body),
	)
	request.Header.Set("Content-Type", "application/json")

	return request
}

func newTestHandler(teams teamService) stdhttp.Handler {
	return newTestHandlerWithServices(teams, nil)
}

func newTestHandlerWithServices(
	teams teamService,
	services serviceService,
) stdhttp.Handler {
	return requestid.Middleware(NewHandler(teams, services))
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
	updateID    *int64
	updateInput *catalog.UpdateTeamInput
	deleteID    *int64
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

func (s *recordingTeamService) UpdateTeam(
	_ context.Context,
	id int64,
	input catalog.UpdateTeamInput,
) (catalog.Team, error) {
	s.updateID = &id
	s.updateInput = &input
	return s.team, s.err
}

func (s *recordingTeamService) DeleteTeam(
	_ context.Context,
	id int64,
) error {
	s.deleteID = &id
	return s.err
}

type recordingServiceService struct {
	createInput *catalog.CreateServiceInput
	getID       *int64
	listInput   *catalog.ListServicesInput
	updateID    *int64
	updateInput *catalog.UpdateServiceInput
	deleteID    *int64
	detail      catalog.ServiceDetail
	service     catalog.Service
	page        catalog.ServicePage
	err         error
}

func (s *recordingServiceService) CreateService(
	_ context.Context,
	input catalog.CreateServiceInput,
) (catalog.ServiceDetail, error) {
	s.createInput = &input
	return s.detail, s.err
}

func (s *recordingServiceService) GetServiceByID(
	_ context.Context,
	id int64,
) (catalog.ServiceDetail, error) {
	s.getID = &id
	return s.detail, s.err
}

func (s *recordingServiceService) ListServices(
	_ context.Context,
	input catalog.ListServicesInput,
) (catalog.ServicePage, error) {
	s.listInput = &input
	return s.page, s.err
}

func (s *recordingServiceService) UpdateService(
	_ context.Context,
	id int64,
	input catalog.UpdateServiceInput,
) (catalog.Service, error) {
	s.updateID = &id
	s.updateInput = &input
	return s.service, s.err
}

func (s *recordingServiceService) DeleteService(
	_ context.Context,
	id int64,
) error {
	s.deleteID = &id
	return s.err
}
