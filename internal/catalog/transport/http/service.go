package cataloghttp

import (
	"context"
	"time"

	"github.com/9AliMay9/lyapus/internal/catalog"
)

type serviceService interface {
	CreateService(
		context.Context,
		catalog.CreateServiceInput,
	) (catalog.ServiceDetail, error)

	GetServiceByID(
		context.Context,
		int64,
	) (catalog.ServiceDetail, error)

	ListServices(
		context.Context,
		catalog.ListServicesInput,
	) (catalog.ServicePage, error)

	UpdateService(
		context.Context,
		int64,
		catalog.UpdateServiceInput,
	) (catalog.Service, error)

	DeleteService(context.Context, int64) error
}

type createServiceRequest struct {
	TeamID       int64                      `json:"team_id"`
	Slug         string                     `json:"slug"`
	Name         string                     `json:"name"`
	Description  *string                    `json:"description"`
	Environments []createEnvironmentRequest `json:"environments"`
}

type updateServiceRequest struct {
	Slug        patchString         `json:"slug"`
	Name        patchString         `json:"name"`
	Description patchNullableString `json:"description"`
}

type createEnvironmentRequest struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type serviceResponse struct {
	ID          int64     `json:"id"`
	TeamID      int64     `json:"team_id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type servicePageResponse struct {
	Items      []serviceResponse `json:"items"`
	NextCursor string            `json:"next_cursor"`
}

type environmentResponse struct {
	ID        int64     `json:"id"`
	ServiceID int64     `json:"service_id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type serviceDetailResponse struct {
	ID           int64                 `json:"id"`
	TeamID       int64                 `json:"team_id"`
	Slug         string                `json:"slug"`
	Name         string                `json:"name"`
	Description  *string               `json:"description"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
	Environments []environmentResponse `json:"environments"`
}

func createServiceInputFromRequest(
	request createServiceRequest,
) catalog.CreateServiceInput {
	environments := make(
		[]catalog.CreateEnvironmentInput,
		len(request.Environments),
	)
	for index, environment := range request.Environments {
		environments[index] = catalog.CreateEnvironmentInput{
			Slug: environment.Slug,
			Name: environment.Name,
		}
	}

	return catalog.CreateServiceInput{
		TeamID:       request.TeamID,
		Slug:         request.Slug,
		Name:         request.Name,
		Description:  request.Description,
		Environments: environments,
	}
}

func updateServiceInputFromRequest(
	request updateServiceRequest,
) catalog.UpdateServiceInput {
	input := catalog.UpdateServiceInput{}

	if request.Slug.set {
		slug := request.Slug.value
		input.Slug = &slug
	}
	if request.Name.set {
		name := request.Name.value
		input.Name = &name
	}
	if request.Description.set {
		input.Description = request.Description.value
		input.DescriptionProvided = true
	}

	return input
}

func serviceResponseFromCatalog(service catalog.Service) serviceResponse {
	return serviceResponse{
		ID:          service.ID,
		TeamID:      service.TeamID,
		Slug:        service.Slug,
		Name:        service.Name,
		Description: service.Description,
		CreatedAt:   service.CreatedAt.UTC(),
		UpdatedAt:   service.UpdatedAt.UTC(),
	}
}

func environmentResponseFromCatalog(
	environment catalog.Environment,
) environmentResponse {
	return environmentResponse{
		ID:        environment.ID,
		ServiceID: environment.ServiceID,
		Slug:      environment.Slug,
		Name:      environment.Name,
		CreatedAt: environment.CreatedAt.UTC(),
		UpdatedAt: environment.UpdatedAt.UTC(),
	}
}

func serviceDetailResponseFromCatalog(
	detail catalog.ServiceDetail,
) serviceDetailResponse {
	environments := make(
		[]environmentResponse,
		len(detail.Environments),
	)
	for index, environment := range detail.Environments {
		environments[index] = environmentResponseFromCatalog(environment)
	}

	service := serviceResponseFromCatalog(detail.Service)

	return serviceDetailResponse{
		ID:           service.ID,
		TeamID:       service.TeamID,
		Slug:         service.Slug,
		Name:         service.Name,
		Description:  service.Description,
		CreatedAt:    service.CreatedAt,
		UpdatedAt:    service.UpdatedAt,
		Environments: environments,
	}
}
