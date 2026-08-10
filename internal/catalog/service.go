package catalog

import (
	"context"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	defaultTeamListLimit int32 = 20
	maxTeamListLimit     int32 = 100
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

func validateTeamID(id int64) error {
	if id < 1 {
		return invalidTeamArgument("team ID must be positive")
	}

	return nil
}

func normalizeTeamSlug(slug string) (string, error) {
	if !teamSlugPattern.MatchString(slug) {
		return "", invalidTeamArgument("slug must match ^[a-z][a-z0-9-]{0,62}$")
	}

	return slug, nil
}

func normalizeTeamName(name string) (string, error) {
	normalized := strings.TrimSpace(name)
	if normalized == "" || utf8.RuneCountInString(normalized) > 100 {
		return "", invalidTeamArgument("name must contain between 1 and 100 characters")
	}

	return normalized, nil
}

func invalidTeamArgument(message string) error {
	return &InvalidArgumentError{
		Message: message,
	}
}
