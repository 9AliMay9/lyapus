package catalog

import "context"

type CreateTeamInput struct {
	Slug string
	Name string
}

type TeamRepository interface {
	CreateTeam(ctx context.Context, input CreateTeamInput) (Team, error)
	GetTeamByID(ctx context.Context, id int64) (Team, error)
}
