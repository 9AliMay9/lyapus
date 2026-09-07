package catalog

import (
	"context"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	defaultTeamListLimit        int32 = 20
	maxTeamListLimit            int32 = 100
	defaultServiceListLimit     int32 = 20
	maxServiceListLimit         int32 = 100
	defaultEnvironmentListLimit int32 = 20
	maxEnvironmentListLimit     int32 = 100
)

var teamSlugPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,62}$`)

type TeamService struct {
	repository TeamRepository
}

func NewTeamService(repository TeamRepository) *TeamService {
	return &TeamService{
		repository: repository,
	}
}

func (s *TeamService) CreateTeam(ctx context.Context, input CreateTeamInput) (Team, error) {
	normalized, err := normalizeCreateTeamInput(input)
	if err != nil {
		return Team{}, err
	}

	return s.repository.CreateTeam(ctx, normalized)
}

func (s *TeamService) GetTeamByID(ctx context.Context, id int64) (Team, error) {
	if err := validateTeamID(id); err != nil {
		return Team{}, err
	}

	return s.repository.GetTeamByID(ctx, id)
}

func (s *TeamService) ListTeams(ctx context.Context, input ListTeamsInput) (TeamPage, error) {
	normalized, err := normalizeListTeamsInput(input)
	if err != nil {
		return TeamPage{}, err
	}

	return s.repository.ListTeams(ctx, normalized)
}

func (s *TeamService) UpdateTeam(ctx context.Context, id int64, input UpdateTeamInput) (Team, error) {
	if err := validateTeamID(id); err != nil {
		return Team{}, err
	}

	normalized, err := normalizeUpdateTeamInput(input)
	if err != nil {
		return Team{}, err
	}

	return s.repository.UpdateTeam(ctx, id, normalized)
}

func (s *TeamService) DeleteTeam(ctx context.Context, id int64) error {
	if err := validateTeamID(id); err != nil {
		return err
	}

	return s.repository.DeleteTeam(ctx, id)
}

type ServiceService struct {
	repository ServiceRepository
}

func NewServiceService(repository ServiceRepository) *ServiceService {
	return &ServiceService{
		repository: repository,
	}
}

func (s *ServiceService) CreateService(
	ctx context.Context,
	input CreateServiceInput,
) (ServiceDetail, error) {
	normalized, err := normalizeCreateServiceInput(input)
	if err != nil {
		return ServiceDetail{}, err
	}

	return s.repository.CreateService(ctx, normalized)
}

func (s *ServiceService) GetServiceByID(
	ctx context.Context,
	id int64,
) (ServiceDetail, error) {
	if err := validateServiceID(id); err != nil {
		return ServiceDetail{}, err
	}

	return s.repository.GetServiceByID(ctx, id)
}

func (s *ServiceService) ListServices(
	ctx context.Context,
	input ListServicesInput,
) (ServicePage, error) {
	normalized, err := normalizeListServicesInput(input)
	if err != nil {
		return ServicePage{}, err
	}

	return s.repository.ListServices(ctx, normalized)
}

func (s *ServiceService) UpdateService(
	ctx context.Context,
	id int64,
	input UpdateServiceInput,
) (Service, error) {
	if err := validateServiceID(id); err != nil {
		return Service{}, err
	}

	normalized, err := normalizeUpdateServiceInput(input)
	if err != nil {
		return Service{}, err
	}

	return s.repository.UpdateService(ctx, id, normalized)
}

func (s *ServiceService) DeleteService(ctx context.Context, id int64) error {
	if err := validateServiceID(id); err != nil {
		return err
	}

	return s.repository.DeleteService(ctx, id)
}

type EnvironmentService struct {
	repository EnvironmentRepository
}

func NewEnvironmentService(
	repository EnvironmentRepository,
) *EnvironmentService {
	return &EnvironmentService{
		repository: repository,
	}
}

func (s *EnvironmentService) CreateEnvironment(
	ctx context.Context,
	input CreateEnvironmentInput,
) (Environment, error) {
	normalized, err := normalizeCreateEnvironmentInput(input)
	if err != nil {
		return Environment{}, err
	}

	return s.repository.CreateEnvironment(ctx, normalized)
}

func (s *EnvironmentService) GetEnvironmentByID(
	ctx context.Context,
	id int64,
) (Environment, error) {
	if err := validateEnvironmentID(id); err != nil {
		return Environment{}, err
	}

	return s.repository.GetEnvironmentByID(ctx, id)
}

func (s *EnvironmentService) ListEnvironments(
	ctx context.Context,
	input ListEnvironmentsInput,
) (EnvironmentPage, error) {
	normalized, err := normalizeListEnvironmentsInput(input)
	if err != nil {
		return EnvironmentPage{}, err
	}

	return s.repository.ListEnvironments(ctx, normalized)
}

func (s *EnvironmentService) UpdateEnvironment(
	ctx context.Context,
	id int64,
	input UpdateEnvironmentInput,
) (Environment, error) {
	if err := validateEnvironmentID(id); err != nil {
		return Environment{}, err
	}

	normalized, err := normalizeUpdateEnvironmentInput(input)
	if err != nil {
		return Environment{}, err
	}

	return s.repository.UpdateEnvironment(ctx, id, normalized)
}

func (s *EnvironmentService) DeleteEnvironment(
	ctx context.Context,
	id int64,
) error {
	if err := validateEnvironmentID(id); err != nil {
		return err
	}

	return s.repository.DeleteEnvironment(ctx, id)
}

func normalizeCreateTeamInput(input CreateTeamInput) (CreateTeamInput, error) {
	slug, err := normalizeTeamSlug(input.Slug)
	if err != nil {
		return CreateTeamInput{}, err
	}

	name, err := normalizeTeamName(input.Name)
	if err != nil {
		return CreateTeamInput{}, err
	}

	return CreateTeamInput{
		Slug: slug,
		Name: name,
	}, nil
}

func normalizeCreateServiceInput(
	input CreateServiceInput,
) (CreateServiceInput, error) {
	if err := validateTeamID(input.TeamID); err != nil {
		return CreateServiceInput{}, err
	}

	slug, err := normalizeServiceSlug(input.Slug)
	if err != nil {
		return CreateServiceInput{}, err
	}

	name, err := normalizeServiceName(input.Name)
	if err != nil {
		return CreateServiceInput{}, err
	}

	description, err := normalizeServiceDescription(input.Description)
	if err != nil {
		return CreateServiceInput{}, err
	}

	environments := make([]CreateInitialEnvironmentInput, len(input.Environments))
	for i, environment := range input.Environments {
		normalizedEnvironment, err := normalizeCreateInitialEnvironmentInput(environment)
		if err != nil {
			return CreateServiceInput{}, err
		}
		environments[i] = normalizedEnvironment
	}

	return CreateServiceInput{
		TeamID:       input.TeamID,
		Slug:         slug,
		Name:         name,
		Description:  description,
		Environments: environments,
	}, nil
}

func normalizeUpdateServiceInput(
	input UpdateServiceInput,
) (UpdateServiceInput, error) {
	if input.Slug == nil && input.Name == nil && !input.DescriptionProvided {
		return UpdateServiceInput{}, invalidTeamArgument(
			"at least one field must be provided",
		)
	}

	var normalized UpdateServiceInput

	if input.Slug != nil {
		slug, err := normalizeServiceSlug(*input.Slug)
		if err != nil {
			return UpdateServiceInput{}, err
		}
		normalized.Slug = &slug
	}

	if input.Name != nil {
		name, err := normalizeServiceName(*input.Name)
		if err != nil {
			return UpdateServiceInput{}, err
		}
		normalized.Name = &name
	}

	if input.DescriptionProvided {
		description, err := normalizeServiceDescription(input.Description)
		if err != nil {
			return UpdateServiceInput{}, err
		}
		normalized.Description = description
		normalized.DescriptionProvided = true
	}

	return normalized, nil
}

func normalizeCreateInitialEnvironmentInput(
	input CreateInitialEnvironmentInput,
) (CreateInitialEnvironmentInput, error) {
	slug, err := normalizeServiceSlug(input.Slug)
	if err != nil {
		return CreateInitialEnvironmentInput{}, err
	}

	name, err := normalizeServiceName(input.Name)
	if err != nil {
		return CreateInitialEnvironmentInput{}, err
	}

	return CreateInitialEnvironmentInput{
		Slug: slug,
		Name: name,
	}, nil
}

func normalizeUpdateTeamInput(input UpdateTeamInput) (UpdateTeamInput, error) {
	if input.Slug == nil && input.Name == nil {
		return UpdateTeamInput{}, invalidTeamArgument("at least one field must be provided")
	}

	var normalized UpdateTeamInput

	if input.Slug != nil {
		slug, err := normalizeTeamSlug(*input.Slug)
		if err != nil {
			return UpdateTeamInput{}, err
		}
		normalized.Slug = &slug
	}

	if input.Name != nil {
		name, err := normalizeTeamName(*input.Name)
		if err != nil {
			return UpdateTeamInput{}, err
		}
		normalized.Name = &name
	}

	return normalized, nil
}

func normalizeListTeamsInput(input ListTeamsInput) (ListTeamsInput, error) {
	switch {
	case input.Limit == 0:
		input.Limit = defaultTeamListLimit
	case input.Limit < 0 || input.Limit > maxTeamListLimit:
		return ListTeamsInput{}, invalidTeamArgument("limit must be between 1 and 100")
	}

	if input.After == nil {
		return input, nil
	}
	if input.After.ID < 1 {
		return ListTeamsInput{}, invalidTeamArgument("cursor ID must be positive")
	}
	if input.After.CreatedAt.IsZero() {
		return ListTeamsInput{}, invalidTeamArgument("cursor created_at must be set")
	}

	after := *input.After
	after.CreatedAt = after.CreatedAt.UTC()
	input.After = &after

	return input, nil
}

func normalizeListServicesInput(
	input ListServicesInput,
) (ListServicesInput, error) {
	if input.TeamID != nil {
		if err := validateTeamID(*input.TeamID); err != nil {
			return ListServicesInput{}, err
		}

		teamID := *input.TeamID
		input.TeamID = &teamID
	}

	switch {
	case input.Limit == 0:
		input.Limit = defaultServiceListLimit
	case input.Limit < 0 || input.Limit > maxServiceListLimit:
		return ListServicesInput{}, invalidTeamArgument("limit must be between 1 and 100")
	}

	if input.After == nil {
		return input, nil
	}
	if input.After.ID < 1 {
		return ListServicesInput{}, invalidTeamArgument("cursor ID must be positive")
	}
	if input.After.CreatedAt.IsZero() {
		return ListServicesInput{}, invalidTeamArgument("cursor created_at must be set")
	}

	after := *input.After
	after.CreatedAt = after.CreatedAt.UTC()
	input.After = &after

	return input, nil
}

func normalizeCreateEnvironmentInput(
	input CreateEnvironmentInput,
) (CreateEnvironmentInput, error) {
	if err := validateServiceID(input.ServiceID); err != nil {
		return CreateEnvironmentInput{}, err
	}

	slug, err := normalizeEnvironmentSlug(input.Slug)
	if err != nil {
		return CreateEnvironmentInput{}, err
	}

	name, err := normalizeEnvironmentName(input.Name)
	if err != nil {
		return CreateEnvironmentInput{}, err
	}

	return CreateEnvironmentInput{
		ServiceID: input.ServiceID,
		Slug:      slug,
		Name:      name,
	}, nil
}

func normalizeUpdateEnvironmentInput(
	input UpdateEnvironmentInput,
) (UpdateEnvironmentInput, error) {
	if input.Slug == nil && input.Name == nil {
		return UpdateEnvironmentInput{}, invalidTeamArgument(
			"at least one field must be provided",
		)
	}

	var normalized UpdateEnvironmentInput

	if input.Slug != nil {
		slug, err := normalizeEnvironmentSlug(*input.Slug)
		if err != nil {
			return UpdateEnvironmentInput{}, err
		}
		normalized.Slug = &slug
	}

	if input.Name != nil {
		name, err := normalizeEnvironmentName(*input.Name)
		if err != nil {
			return UpdateEnvironmentInput{}, err
		}
		normalized.Name = &name
	}

	return normalized, nil
}

func normalizeListEnvironmentsInput(
	input ListEnvironmentsInput,
) (ListEnvironmentsInput, error) {
	if input.ServiceID != nil {
		if err := validateServiceID(*input.ServiceID); err != nil {
			return ListEnvironmentsInput{}, err
		}

		serviceID := *input.ServiceID
		input.ServiceID = &serviceID
	}

	switch {
	case input.Limit == 0:
		input.Limit = defaultEnvironmentListLimit
	case input.Limit < 0 || input.Limit > maxEnvironmentListLimit:
		return ListEnvironmentsInput{}, invalidTeamArgument(
			"limit must be between 1 and 100",
		)
	}

	if input.After == nil {
		return input, nil
	}
	if input.After.ID < 1 {
		return ListEnvironmentsInput{}, invalidTeamArgument(
			"cursor ID must be positive",
		)
	}
	if input.After.CreatedAt.IsZero() {
		return ListEnvironmentsInput{}, invalidTeamArgument(
			"cursor created_at must be set",
		)
	}

	after := *input.After
	after.CreatedAt = after.CreatedAt.UTC()
	input.After = &after

	return input, nil
}

func validateTeamID(id int64) error {
	if id < 1 {
		return invalidTeamArgument("team ID must be positive")
	}

	return nil
}

func validateServiceID(id int64) error {
	if id < 1 {
		return invalidTeamArgument("service ID must be positive")
	}

	return nil
}

func validateEnvironmentID(id int64) error {
	if id < 1 {
		return invalidTeamArgument("environment ID must be positive")
	}

	return nil
}

func normalizeTeamSlug(slug string) (string, error) {
	if !teamSlugPattern.MatchString(slug) {
		return "", invalidTeamArgument("slug must match ^[a-z][a-z0-9-]{0,62}$")
	}

	return slug, nil
}

func normalizeServiceSlug(slug string) (string, error) {
	return normalizeTeamSlug(slug)
}

func normalizeTeamName(name string) (string, error) {
	normalized := strings.TrimSpace(name)
	if normalized == "" || utf8.RuneCountInString(normalized) > 100 {
		return "", invalidTeamArgument("name must contain between 1 and 100 characters")
	}

	return normalized, nil
}

func normalizeServiceName(name string) (string, error) {
	return normalizeTeamName(name)
}

func normalizeEnvironmentSlug(slug string) (string, error) {
	return normalizeTeamSlug(slug)
}

func normalizeEnvironmentName(name string) (string, error) {
	return normalizeTeamName(name)
}

func normalizeServiceDescription(
	description *string,
) (*string, error) {
	if description == nil {
		return nil, nil
	}
	if utf8.RuneCountInString(*description) > 500 {
		return nil, invalidTeamArgument(
			"description must contain at most 500 characters",
		)
	}

	normalized := *description
	return &normalized, nil
}

func invalidTeamArgument(message string) error {
	return &InvalidArgumentError{
		Message: message,
	}
}
