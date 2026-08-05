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
