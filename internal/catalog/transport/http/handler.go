package cataloghttp

import (
	"context"
	stdhttp "net/http"
	"time"

	"github.com/9AliMay9/lyapus/internal/catalog"
	"github.com/9AliMay9/lyapus/internal/platform/requestid"
	"github.com/go-chi/chi/v5"
)

type teamService interface {
	CreateTeam(context.Context, catalog.CreateTeamInput) (catalog.Team, error)
}

type Handler struct {
	teams teamService
}

type createTeamRequest struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type teamResponse struct {
	ID        int64     `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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

func teamResponseFromCatalog(team catalog.Team) teamResponse {
	return teamResponse{
		ID:        team.ID,
		Slug:      team.Slug,
		Name:      team.Name,
		CreatedAt: team.CreatedAt.UTC(),
		UpdatedAt: team.UpdatedAt.UTC(),
	}
}
