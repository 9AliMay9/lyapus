package catalog

import "context"

type CreateTeamInput struct {
	Slug string
	Name string
}

type TeamRepository interface {
	CreateTeam(ctx context.Context, input CreateTeamInput) (Team, error)
	GetTeamByID(ctx context.Context, id int64) (Team, error)
	ListTeams(ctx context.Context, input ListTeamsInput) (TeamPage, error)
	UpdateTeam(ctx context.Context, id int64, input UpdateTeamInput) (Team, error)
	DeleteTeam(ctx context.Context, id int64) error
}

type ServiceRepository interface {
	CreateService(ctx context.Context, input CreateServiceInput) (ServiceDetail, error)
	GetServiceByID(ctx context.Context, id int64) (ServiceDetail, error)
	ListServices(ctx context.Context, input ListServicesInput) (ServicePage, error)
	UpdateService(ctx context.Context, id int64, input UpdateServiceInput) (Service, error)
	DeleteService(ctx context.Context, id int64) error
}
