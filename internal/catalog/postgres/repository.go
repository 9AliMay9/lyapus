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

type serviceDB interface {
	sqlcgen.DBTX
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}

type ServiceRepository struct {
	db      serviceDB
	queries *sqlcgen.Queries
}

var (
	_ catalog.TeamRepository    = (*TeamRepository)(nil)
	_ catalog.ServiceRepository = (*ServiceRepository)(nil)
)

func NewTeamRepository(db sqlcgen.DBTX) *TeamRepository {
	return &TeamRepository{
		queries: sqlcgen.New(db),
	}
}

func NewServiceRepository(db serviceDB) *ServiceRepository {
	return &ServiceRepository{
		db:      db,
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
		Slug: textFromStringPointer(input.Slug),
		Name: textFromStringPointer(input.Name),
		ID:   id,
	})
	if err != nil {
		return catalog.Team{}, classifyTeamError("update team", err)
	}

	return teamFromRow(row)
}

func textFromStringPointer(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}

	return pgtype.Text{
		String: *value,
		Valid:  true,
	}
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

func (r *ServiceRepository) CreateService(
	ctx context.Context,
	input catalog.CreateServiceInput,
) (catalog.ServiceDetail, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return catalog.ServiceDetail{}, classifyServiceError(
			"begin create service transaction",
			err,
		)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	queries := r.queries.WithTx(tx)

	serviceRow, err := queries.CreateService(ctx, sqlcgen.CreateServiceParams{
		TeamID:      input.TeamID,
		Slug:        input.Slug,
		Name:        input.Name,
		Description: textFromStringPointer(input.Description),
	})
	if err != nil {
		return catalog.ServiceDetail{}, classifyServiceError("create service", err)
	}

	service, err := serviceFromRow(serviceRow)
	if err != nil {
		return catalog.ServiceDetail{}, err
	}

	environments := make([]catalog.Environment, 0, len(input.Environments))
	for _, inputEnvironment := range input.Environments {
		environmentRow, err := queries.CreateEnvironment(
			ctx,
			sqlcgen.CreateEnvironmentParams{
				ServiceID: service.ID,
				Slug:      inputEnvironment.Slug,
				Name:      inputEnvironment.Name,
			},
		)
		if err != nil {
			return catalog.ServiceDetail{}, classifyServiceError(
				"create initial environment",
				err,
			)
		}

		environment, err := environmentFromRow(environmentRow)
		if err != nil {
			return catalog.ServiceDetail{}, err
		}
		environments = append(environments, environment)
	}

	if err := tx.Commit(ctx); err != nil {
		return catalog.ServiceDetail{}, classifyServiceError(
			"commit create service transaction",
			err,
		)
	}

	return catalog.ServiceDetail{
		Service:      service,
		Environments: environments,
	}, nil
}

func (r *ServiceRepository) GetServiceByID(
	ctx context.Context,
	id int64,
) (catalog.ServiceDetail, error) {
	serviceRow, err := r.queries.GetServiceByID(ctx, id)
	if err != nil {
		return catalog.ServiceDetail{}, classifyServiceError("get service by ID", err)
	}

	service, err := serviceFromRow(serviceRow)
	if err != nil {
		return catalog.ServiceDetail{}, err
	}

	environmentRows, err := r.queries.ListEnvironmentsByServiceID(ctx, id)
	if err != nil {
		return catalog.ServiceDetail{}, fmt.Errorf(
			"list service environments: %w",
			err,
		)
	}

	environments := make([]catalog.Environment, 0, len(environmentRows))
	for _, row := range environmentRows {
		environment, err := environmentFromRow(row)
		if err != nil {
			return catalog.ServiceDetail{}, err
		}
		environments = append(environments, environment)
	}

	return catalog.ServiceDetail{
		Service:      service,
		Environments: environments,
	}, nil
}

func (r *ServiceRepository) ListServices(
	ctx context.Context,
	input catalog.ListServicesInput,
) (catalog.ServicePage, error) {
	limit, err := serviceListLimitWithExtra(input.Limit)
	if err != nil {
		return catalog.ServicePage{}, err
	}

	var rows []sqlcgen.Service

	switch {
	case input.TeamID == nil && input.After == nil:

		rows, err = r.queries.ListServicesFirstPage(ctx, limit)

	case input.TeamID != nil && input.After == nil:
		rows, err = r.queries.ListServicesFirstPageByTeamID(
			ctx,
			sqlcgen.ListServicesFirstPageByTeamIDParams{
				TeamID: *input.TeamID,
				Limit:  limit,
			},
		)

	case input.TeamID == nil:
		rows, err = r.queries.ListServicesAfterCursor(
			ctx,
			sqlcgen.ListServicesAfterCursorParams{
				CreatedAt: pgtype.Timestamptz{
					Time:  input.After.CreatedAt.UTC(),
					Valid: true,
				},
				ID:    input.After.ID,
				Limit: limit,
			},
		)

	default:
		rows, err = r.queries.ListServicesAfterCursorByTeamID(
			ctx,
			sqlcgen.ListServicesAfterCursorByTeamIDParams{
				TeamID: *input.TeamID,
				CreatedAt: pgtype.Timestamptz{
					Time:  input.After.CreatedAt.UTC(),
					Valid: true,
				},
				ID:    input.After.ID,
				Limit: limit,
			},
		)
	}
	if err != nil {
		return catalog.ServicePage{}, fmt.Errorf("list services: %w", err)
	}

	return servicePageFromRows(rows, input.Limit)
}

func serviceListLimitWithExtra(limit int32) (int32, error) {
	if limit < 1 || limit == maxInt32 {
		return 0, catalog.ErrInvalidArgument
	}

	return limit + 1, nil
}

func servicePageFromRows(
	rows []sqlcgen.Service,
	limit int32,
) (catalog.ServicePage, error) {
	services := make([]catalog.Service, 0, len(rows))
	for _, row := range rows {
		service, err := serviceFromRow(row)
		if err != nil {
			return catalog.ServicePage{}, err
		}
		services = append(services, service)
	}

	if len(services) <= int(limit) {
		return catalog.ServicePage{Services: services}, nil
	}

	last := services[int(limit)-1]
	next := catalog.ServiceCursor{
		CreatedAt: last.CreatedAt,
		ID:        last.ID,
	}

	return catalog.ServicePage{
		Services: services[:int(limit)],
		Next:     &next,
	}, nil
}

func serviceFromRow(row sqlcgen.Service) (catalog.Service, error) {
	if !row.CreatedAt.Valid {
		return catalog.Service{}, fmt.Errorf("map service: created_at is null")
	}
	if !row.UpdatedAt.Valid {
		return catalog.Service{}, fmt.Errorf("map service: updated_at is null")
	}

	var description *string
	if row.Description.Valid {
		value := row.Description.String
		description = &value
	}

	return catalog.Service{
		ID:          row.ID,
		TeamID:      row.TeamID,
		Slug:        row.Slug,
		Name:        row.Name,
		Description: description,
		CreatedAt:   row.CreatedAt.Time.UTC(),
		UpdatedAt:   row.UpdatedAt.Time.UTC(),
	}, nil
}

func environmentFromRow(row sqlcgen.Environment) (catalog.Environment, error) {
	if !row.CreatedAt.Valid {
		return catalog.Environment{}, fmt.Errorf("map environment: created_at is null")
	}
	if !row.UpdatedAt.Valid {
		return catalog.Environment{}, fmt.Errorf("map environment: updated_at is null")
	}

	return catalog.Environment{
		ID:        row.ID,
		ServiceID: row.ServiceID,
		Slug:      row.Slug,
		Name:      row.Name,
		CreatedAt: row.CreatedAt.Time.UTC(),
		UpdatedAt: row.UpdatedAt.Time.UTC(),
	}, nil
}

func classifyServiceError(operation string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return catalog.ErrNotFound
	}

	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "23505":
			return catalog.ErrConflict

		case "23503":
			if postgresError.ConstraintName == "services_team_id_fkey" {
				return catalog.ErrNotFound
			}
			return catalog.ErrConflict
		}
	}

	return fmt.Errorf("%s: %w", operation, err)
}
