package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/9AliMay9/lyapus/internal/catalog"
	"github.com/9AliMay9/lyapus/internal/catalog/postgres/sqlcgen"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

const maxInt32 = int32(1<<31 - 1)

type TeamRepository struct {
	queries *sqlcgen.Queries
}

var _ catalog.TeamRepository = (*TeamRepository)(nil)

func NewTeamRepository(db sqlcgen.DBTX) *TeamRepository {
	return &TeamRepository{
		queries: sqlcgen.New(db),
	}
}

func (r *TeamRepository) CreateTeam(ctx context.Context, input catalog.CreateTeamInput) (catalog.Team, error) {
	row, err := r.queries.CreateTeam(ctx, sqlcgen.CreateTeamParams{
		Slug: input.Slug,
		Name: input.Name,
	})
	if err != nil {
		return catalog.Team{}, classifyTeamError("create team", err)
	}

	return teamFromRow(row)
}

func (r *TeamRepository) GetTeamByID(ctx context.Context, id int64) (catalog.Team, error) {
	row, err := r.queries.GetTeamByID(ctx, id)
	if err != nil {
		return catalog.Team{}, classifyTeamError("get team by ID", err)
	}

	return teamFromRow(row)
}

func (r *TeamRepository) ListTeams(ctx context.Context, input catalog.ListTeamsInput) (catalog.TeamPage, error) {
	limit, err := teamListLimitWithExtra(input.Limit)
	if err != nil {
		return catalog.TeamPage{}, err
	}

	var rows []sqlcgen.Team
	if input.After == nil {
		rows, err = r.queries.ListTeamsFirstPage(ctx, limit)
	} else {
		rows, err = r.queries.ListTeamsAfterCursor(ctx, sqlcgen.ListTeamsAfterCursorParams{
			CreatedAt: pgtype.Timestamptz{
				Time:  input.After.CreatedAt.UTC(),
				Valid: true,
			},
			ID:    input.After.ID,
			Limit: limit,
		})
	}
	if err != nil {
		return catalog.TeamPage{}, fmt.Errorf("list teams: %w", err)
	}

	return teamPageFromRows(rows, input.Limit)
}

func (r *TeamRepository) UpdateTeam(ctx context.Context, id int64, input catalog.UpdateTeamInput) (catalog.Team, error) {
	row, err := r.queries.UpdateTeam(ctx, sqlcgen.UpdateTeamParams{
		ID:   id,
		Slug: input.Slug,
		Name: input.Name,
	})
	if err != nil {
		return catalog.Team{}, classifyTeamError("update team", err)
	}

	return teamFromRow(row)
}

func (r *TeamRepository) DeleteTeam(ctx context.Context, id int64) error {
	_, err := r.queries.DeleteTeam(ctx, id)
	if err != nil {
		return classifyTeamError("delete team", err)
	}

	return nil
}

func teamListLimitWithExtra(limit int32) (int32, error) {
	if limit < 1 || limit == maxInt32 {
		return 0, catalog.ErrInvalidArgument
	}

	return limit + 1, nil
}

func teamPageFromRows(rows []sqlcgen.Team, limit int32) (catalog.TeamPage, error) {
	teams := make([]catalog.Team, 0, len(rows))
	for _, row := range rows {
		team, err := teamFromRow(row)
		if err != nil {
			return catalog.TeamPage{}, err
		}
		teams = append(teams, team)
	}

	if len(teams) <= int(limit) {
		return catalog.TeamPage{Teams: teams}, nil
	}

	last := teams[int(limit)-1]
	next := catalog.TeamCursor{
		CreatedAt: last.CreatedAt,
		ID:        last.ID,
	}

	return catalog.TeamPage{
		Teams: teams[:int(limit)],
		Next:  &next,
	}, nil
}

func teamFromRow(row sqlcgen.Team) (catalog.Team, error) {
	if !row.CreatedAt.Valid {
		return catalog.Team{}, fmt.Errorf("map team: created_at is null")
	}
	if !row.UpdatedAt.Valid {
		return catalog.Team{}, fmt.Errorf("map team: updated_at is null")
	}

	return catalog.Team{
		ID:        row.ID,
		Slug:      row.Slug,
		Name:      row.Name,
		CreatedAt: row.CreatedAt.Time.UTC(),
		UpdatedAt: row.UpdatedAt.Time.UTC(),
	}, nil
}

func classifyTeamError(operation string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return catalog.ErrNotFound
	}

	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "23503", "23505":
			return catalog.ErrConflict
		}
	}

	return fmt.Errorf("%s: %w", operation, err)
}
