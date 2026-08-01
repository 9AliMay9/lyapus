package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/9AliMay9/lyapus/internal/catalog"
	"github.com/9AliMay9/lyapus/internal/catalog/postgres/sqlcgen"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

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
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return catalog.ErrConflict
	}

	return fmt.Errorf("%s: %w", operation, err)
}
