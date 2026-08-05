package catalog

import "time"

type Team struct {
	ID        int64
	Slug      string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TeamCursor struct {
	CreatedAt time.Time
	ID        int64
}

type ListTeamsInput struct {
	Limit int32
	After *TeamCursor
}

type TeamPage struct {
	Teams []Team
	Next  *TeamCursor
}

type UpdateTeamInput struct {
	Slug string
	Name string
}
