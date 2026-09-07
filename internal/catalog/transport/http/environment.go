package cataloghttp

import (
	"context"
	"time"

	"github.com/9AliMay9/lyapus/internal/catalog"
)

type environmentService interface {
	CreateEnvironment(context.Context, catalog.CreateEnvironmentInput) (catalog.Environment, error)
	GetEnvironmentByID(context.Context, int64) (catalog.Environment, error)
	ListEnvironments(context.Context, catalog.ListEnvironmentsInput) (catalog.EnvironmentPage, error)
	UpdateEnvironment(context.Context, int64, catalog.UpdateEnvironmentInput) (catalog.Environment, error)
	DeleteEnvironment(context.Context, int64) error
}

type createEnvironmentRequest struct {
	ServiceID int64  `json:"service_id"`
	Slug      string `json:"slug"`
	Name      string `json:"name"`
}

type updateEnvironmentRequest struct {
	Slug patchString `json:"slug"`
	Name patchString `json:"name"`
}

type environmentResponse struct {
	ID        int64     `json:"id"`
	ServiceID int64     `json:"service_id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type environmentPageResponse struct {
	Items      []environmentResponse `json:"items"`
	NextCursor string                `json:"next_cursor"`
}

func createEnvironmentInputFromRequest(
	request createEnvironmentRequest,
) catalog.CreateEnvironmentInput {
	return catalog.CreateEnvironmentInput{
		ServiceID: request.ServiceID,
		Slug:      request.Slug,
		Name:      request.Name,
	}
}

func updateEnvironmentInputFromRequest(
	request updateEnvironmentRequest,
) catalog.UpdateEnvironmentInput {
	input := catalog.UpdateEnvironmentInput{}

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

func environmentResponseFromCatalog(environment catalog.Environment) environmentResponse {
	return environmentResponse{
		ID:        environment.ID,
		ServiceID: environment.ServiceID,
		Slug:      environment.Slug,
		Name:      environment.Name,
		CreatedAt: environment.CreatedAt.UTC(),
		UpdatedAt: environment.UpdatedAt.UTC(),
	}
}
