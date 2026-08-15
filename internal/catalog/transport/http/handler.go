package cataloghttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	stdhttp "net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/9AliMay9/lyapus/internal/catalog"
	"github.com/9AliMay9/lyapus/internal/platform/requestid"
	"github.com/go-chi/chi/v5"
)

type teamService interface {
	CreateTeam(context.Context, catalog.CreateTeamInput) (catalog.Team, error)
	GetTeamByID(context.Context, int64) (catalog.Team, error)
	ListTeams(context.Context, catalog.ListTeamsInput) (catalog.TeamPage, error)
	UpdateTeam(context.Context, int64, catalog.UpdateTeamInput) (catalog.Team, error)
	DeleteTeam(context.Context, int64) error
}

type Handler struct {
	teams teamService
}

type createTeamRequest struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type updateTeamRequest struct {
	Slug patchString `json:"slug"`
	Name patchString `json:"name"`
}

type patchString struct {
	value string
	set   bool
}

type teamResponse struct {
	ID        int64     `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type teamPageResponse struct {
	Items      []teamResponse `json:"items"`
	NextCursor string         `json:"next_cursor"`
}

func NewHandler(teams teamService) stdhttp.Handler {
	handler := Handler{
		teams: teams,
	}

	router := chi.NewRouter()

	router.NotFound(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		writeError(
			w,
			stdhttp.StatusNotFound,
			"not_found",
			"resource not found",
			requestid.FromContext(r.Context()),
		)
	})
	router.MethodNotAllowed(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		writeError(
			w,
			stdhttp.StatusMethodNotAllowed,
			"method_not_allowed",
			"method not allowed",
			requestid.FromContext(r.Context()),
		)
	})

	router.Post("/v1/teams", handler.createTeam)
	router.Get("/v1/teams", handler.listTeams)
	router.Get("/v1/teams/{team_id}", handler.getTeam)
	router.Patch("/v1/teams/{team_id}", handler.updateTeam)
	router.Delete("/v1/teams/{team_id}", handler.deleteTeam)

	return router
}

func (h Handler) createTeam(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var request createTeamRequest
	if err := decodeJSONBody(w, r, &request); err != nil {
		writeError(
			w,
			stdhttp.StatusBadRequest,
			"invalid_argument",
			"invalid request body",
			requestid.FromContext(r.Context()),
		)
		return
	}

	team, err := h.teams.CreateTeam(r.Context(), catalog.CreateTeamInput{
		Slug: request.Slug,
		Name: request.Name,
	})
	if err != nil {
		writeCatalogError(w, err, requestid.FromContext(r.Context()))
		return
	}

	writeJSON(w, stdhttp.StatusCreated, teamResponseFromCatalog(team))
}

func (h Handler) getTeam(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	id, err := parsePositiveTeamID(chi.URLParam(r, "team_id"))
	if err != nil {
		writeCatalogError(w, err, requestid.FromContext(r.Context()))
		return
	}

	team, err := h.teams.GetTeamByID(r.Context(), id)
	if err != nil {
		writeCatalogError(w, err, requestid.FromContext(r.Context()))
		return
	}

	writeJSON(w, stdhttp.StatusOK, teamResponseFromCatalog(team))
}

func (h Handler) updateTeam(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	id, err := parsePositiveTeamID(chi.URLParam(r, "team_id"))
	if err != nil {
		writeCatalogError(w, err, requestid.FromContext(r.Context()))
		return
	}

	var request updateTeamRequest
	if err := decodeJSONBody(w, r, &request); err != nil {
		writeError(
			w,
			stdhttp.StatusBadRequest,
			"invalid_argument",
			"invalid request body",
			requestid.FromContext(r.Context()),
		)
		return
	}

	team, err := h.teams.UpdateTeam(
		r.Context(),
		id,
		updateTeamInputFromRequest(request),
	)
	if err != nil {
		writeCatalogError(w, err, requestid.FromContext(r.Context()))
		return
	}

	writeJSON(w, stdhttp.StatusOK, teamResponseFromCatalog(team))
}

func (h Handler) deleteTeam(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	id, err := parsePositiveTeamID(chi.URLParam(r, "team_id"))
	if err != nil {
		writeCatalogError(w, err, requestid.FromContext(r.Context()))
		return
	}

	if err := h.teams.DeleteTeam(r.Context(), id); err != nil {
		writeCatalogError(w, err, requestid.FromContext(r.Context()))
		return
	}

	w.WriteHeader(stdhttp.StatusNoContent)
}

func (h Handler) listTeams(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	input, err := parseListTeamsInput(r)
	if err != nil {
		writeCatalogError(w, err, requestid.FromContext(r.Context()))
		return
	}

	page, err := h.teams.ListTeams(r.Context(), input)
	if err != nil {
		writeCatalogError(w, err, requestid.FromContext(r.Context()))
		return
	}

	response, err := teamPageResponseFromCatalog(page)
	if err != nil {
		writeError(
			w,
			stdhttp.StatusInternalServerError,
			"internal",
			"internal server error",
			requestid.FromContext(r.Context()),
		)
		return
	}

	writeJSON(w, stdhttp.StatusOK, response)
}

func parseListTeamsInput(r *stdhttp.Request) (catalog.ListTeamsInput, error) {
	input := catalog.ListTeamsInput{}
	query := r.URL.Query()

	limit, hasLimit, err := singleQueryValue(query, "limit")
	if err != nil {
		return catalog.ListTeamsInput{}, err
	}
	if hasLimit {
		parsed, err := strconv.ParseInt(limit, 10, 64)
		if err != nil {
			message := "limit must be an integer"
			if errors.Is(err, strconv.ErrRange) {
				message = "limit must be between 1 and 100"
			}
			return catalog.ListTeamsInput{}, &catalog.InvalidArgumentError{
				Message: message,
			}
		}
		if parsed < 1 || parsed > 100 {
			return catalog.ListTeamsInput{}, &catalog.InvalidArgumentError{
				Message: "limit must be between 1 and 100",
			}
		}
		input.Limit = int32(parsed)
	}

	cursor, hasCursor, err := singleQueryValue(query, "cursor")
	if err != nil {
		return catalog.ListTeamsInput{}, err
	}
	if hasCursor {
		after, err := decodeTeamCursor(cursor)
		if err != nil {
			return catalog.ListTeamsInput{}, err
		}
		input.After = &after
	}

	return input, nil
}

func singleQueryValue(query url.Values, name string) (string, bool, error) {
	values, present := query[name]
	if !present {
		return "", false, nil
	}
	if len(values) != 1 {
		return "", true, &catalog.InvalidArgumentError{
			Message: name + " must be provided once",
		}
	}
	return values[0], true, nil
}

func teamPageResponseFromCatalog(page catalog.TeamPage) (teamPageResponse, error) {
	items := make([]teamResponse, len(page.Teams))
	for i, team := range page.Teams {
		items[i] = teamResponseFromCatalog(team)
	}

	response := teamPageResponse{
		Items: items,
	}

	if page.Next == nil {
		return response, nil
	}

	cursor, err := encodeTeamCursor(*page.Next)
	if err != nil {
		return teamPageResponse{}, fmt.Errorf("encode next team cursor: %w", err)
	}

	response.NextCursor = cursor
	return response, nil
}

func parsePositiveTeamID(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id < 1 {
		return 0, &catalog.InvalidArgumentError{
			Message: "team ID must be a positive integer",
		}
	}

	return id, nil
}

func teamResponseFromCatalog(team catalog.Team) teamResponse {
	return teamResponse{
		ID:        team.ID,
		Slug:      team.Slug,
		Name:      team.Name,
		CreatedAt: team.CreatedAt.UTC(),
		UpdatedAt: team.UpdatedAt.UTC(),
	}
}

func (value *patchString) UnmarshalJSON(data []byte) error {
	value.set = true

	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return errors.New("string field must not be null")
	}

	return json.Unmarshal(data, &value.value)
}

func updateTeamInputFromRequest(request updateTeamRequest) catalog.UpdateTeamInput {
	input := catalog.UpdateTeamInput{}

	if request.Slug.set {
		slug := request.Slug.value
		input.Slug = &slug
	}
	if request.Name.set {
		name := request.Name.value
		input.Name = &name
	}

	return input
}
