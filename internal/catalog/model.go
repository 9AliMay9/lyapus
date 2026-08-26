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
	Slug *string
	Name *string
}

type Service struct {
	ID          int64
	TeamID      int64
	Slug        string
	Name        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Environment struct {
	ID        int64
	ServiceID int64
	Slug      string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ServiceDetail struct {
	Service      Service
	Environments []Environment
}

type ServiceCursor struct {
	CreatedAt time.Time
	ID        int64
}

type ListServicesInput struct {
	TeamID *int64
	Limit  int32
	After  *ServiceCursor
}

type ServicePage struct {
	Services []Service
	Next     *ServiceCursor
}

type CreateEnvironmentInput struct {
	Slug string
	Name string
}

type CreateServiceInput struct {
	TeamID       int64
	Slug         string
	Name         string
	Description  *string
	Environments []CreateEnvironmentInput
}

type UpdateServiceInput struct {
	Slug                *string
	Name                *string
	Description         *string
	DescriptionProvided bool
}
