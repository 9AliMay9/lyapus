package postgres

import (
	"errors"
	"testing"
	"time"

	"github.com/9AliMay9/lyapus/internal/catalog"
	"github.com/9AliMay9/lyapus/internal/catalog/postgres/sqlcgen"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestTeamFromRowMapsValuesToCatalogTeam(t *testing.T) {
	createdAt := time.Date(2026, time.July, 31, 12, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	updatedAt := createdAt.Add(time.Hour)

	row := sqlcgen.Team{
		ID:   42,
		Slug: "platform",
		Name: "Platform",
		CreatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  updatedAt,
			Valid: true,
		},
	}

	got, err := teamFromRow(row)
	if err != nil {
		t.Fatalf("teamFromRow() error = %v", err)
	}

	want := catalog.Team{
		ID:        42,
		Slug:      "platform",
		Name:      "Platform",
		CreatedAt: createdAt.UTC(),
		UpdatedAt: updatedAt.UTC(),
	}
	if got != want {
		t.Fatalf("teamFromRow() = %#v, want %#v", got, want)
	}
}

func TestTeamFromRowRejectsNullTimestamp(t *testing.T) {
	validTimestamp := pgtype.Timestamptz{
		Time:  time.Date(2026, time.July, 31, 0, 0, 0, 0, time.UTC),
		Valid: true,
	}

	tests := []struct {
		name      string
		row       sqlcgen.Team
		wantError string
	}{
		{
			name: "created at",
			row: sqlcgen.Team{
				UpdatedAt: validTimestamp,
			},
			wantError: "map team: created_at is null",
		},
		{
			name: "updated at",
			row: sqlcgen.Team{
				CreatedAt: validTimestamp,
			},
			wantError: "map team: updated_at is null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := teamFromRow(tt.row)

			if err == nil {
				t.Fatal("teamFromRow() error = nil, want an error")
			}
			if err.Error() != tt.wantError {
				t.Fatalf("teamFromRow() error = %q, want %q", err, tt.wantError)
			}
		})
	}
}

func TestTeamListLimitWithExtra(t *testing.T) {
	tests := []struct {
		name    string
		limit   int32
		want    int32
		wantErr error
	}{
		{
			name:  "one",
			limit: 1,
			want:  2,
		},
		{
			name:  "regular limit",
			limit: 20,
			want:  21,
		},
		{
			name:    "zero",
			limit:   0,
			wantErr: catalog.ErrInvalidArgument,
		},
		{
			name:    "negative",
			limit:   -1,
			wantErr: catalog.ErrInvalidArgument,
		},
		{
			name:    "maximum int32",
			limit:   maxInt32,
			wantErr: catalog.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := teamListLimitWithExtra(tt.limit)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("teamListLimitWithExtra() error = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("teamListLimitWithExtra() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestTeamPageFromRows(t *testing.T) {
	base := time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)
	rows := []sqlcgen.Team{
		testSQLCTeam(3, "three", base.Add(2*time.Minute)),
		testSQLCTeam(2, "two", base.Add(time.Minute)),
		testSQLCTeam(1, "one", base),
	}

	page, err := teamPageFromRows(rows, 2)
	if err != nil {
		t.Fatalf("teamPageFromRows() error = %v", err)
	}

	if len(page.Teams) != 2 {
		t.Fatalf("len(Teams) = %d, want 2", len(page.Teams))
	}
	if page.Teams[0].ID != 3 || page.Teams[1].ID != 2 {
		t.Fatalf("Teams IDs = [%d %d], want [3 2]", page.Teams[0].ID, page.Teams[1].ID)
	}

	wantNext := catalog.TeamCursor{
		CreatedAt: base.Add(time.Minute),
		ID:        2,
	}
	if page.Next == nil {
		t.Fatal("Next = nil, want cursor")
	}
	if *page.Next != wantNext {
		t.Fatalf("Next = %#v, want %#v", *page.Next, wantNext)
	}
}

func TestTeamPageFromRowsWithoutExtraRow(t *testing.T) {
	base := time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)
	rows := []sqlcgen.Team{
		testSQLCTeam(2, "two", base.Add(time.Minute)),
		testSQLCTeam(1, "one", base),
	}

	page, err := teamPageFromRows(rows, 2)
	if err != nil {
		t.Fatalf("teamPageFromRows() error = %v", err)
	}

	if len(page.Teams) != 2 {
		t.Fatalf("len(Teams) = %d, want 2", len(page.Teams))
	}
	if page.Next != nil {
		t.Fatalf("Next = %#v, want nil", page.Next)
	}
}

func TestTextFromStringPointer(t *testing.T) {
	value := "platform"

	tests := []struct {
		name  string
		value *string
		want  pgtype.Text
	}{
		{
			name: "nil",
			want: pgtype.Text{},
		},
		{
			name:  "value",
			value: &value,
			want: pgtype.Text{
				String: "platform",
				Valid:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := textFromStringPointer(tt.value)

			if got != tt.want {
				t.Fatalf("textFromStringPointer() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestClassifyTeamError(t *testing.T) {
	unknownError := errors.New("connection lost")

	tests := []struct {
		name      string
		operation string
		err       error
		want      error
	}{
		{
			name:      "not found",
			operation: "get team by ID",
			err:       pgx.ErrNoRows,
			want:      catalog.ErrNotFound,
		},
		{
			name:      "unique violation",
			operation: "create team",
			err: &pgconn.PgError{
				Code: "23505",
			},
			want: catalog.ErrConflict,
		},
		{
			name:      "foreign key violation",
			operation: "delete team",
			err: &pgconn.PgError{
				Code: "23503",
			},
			want: catalog.ErrConflict,
		},
		{
			name:      "unknown error",
			operation: "create team",
			err:       unknownError,
			want:      unknownError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyTeamError(tt.operation, tt.err)

			if !errors.Is(got, tt.want) {
				t.Fatalf("classifyTeamError() = %v, want error matching %v", got, tt.want)
			}
		})
	}
}

func TestServiceFromRowMapsValuesAndNullableDescription(t *testing.T) {
	createdAt := time.Date(2026, time.August, 18, 12, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	updatedAt := createdAt.Add(time.Hour)

	row := sqlcgen.Service{
		ID:     42,
		TeamID: 7,
		Slug:   "catalog-api",
		Name:   "Catalog API",
		Description: pgtype.Text{
			String: "Service catalog API",
			Valid:  true,
		},
		CreatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  updatedAt,
			Valid: true,
		},
	}

	got, err := serviceFromRow(row)
	if err != nil {
		t.Fatalf("serviceFromRow() error = %v", err)
	}

	if got.ID != 42 || got.TeamID != 7 {
		t.Fatalf("serviceFromRow() IDs = (%d, %d), want (42, 7)", got.ID, got.TeamID)
	}
	if got.Slug != "catalog-api" || got.Name != "Catalog API" {
		t.Fatalf("serviceFromRow() = %#v, want catalog-api / Catalog API", got)
	}
	if got.Description == nil || *got.Description != "Service catalog API" {
		t.Fatalf("serviceFromRow() Description = %#v, want Service catalog API", got.Description)
	}
	if !got.CreatedAt.Equal(createdAt.UTC()) {
		t.Fatalf("serviceFromRow() CreatedAt = %s, want %s", got.CreatedAt, createdAt.UTC())
	}
	if !got.UpdatedAt.Equal(updatedAt.UTC()) {
		t.Fatalf("serviceFromRow() UpdatedAt = %s, want %s", got.UpdatedAt, updatedAt.UTC())
	}

	row.Description = pgtype.Text{}

	got, err = serviceFromRow(row)
	if err != nil {
		t.Fatalf("serviceFromRow() with null description error = %v", err)
	}
	if got.Description != nil {
		t.Fatalf("serviceFromRow() null Description = %#v, want nil", got.Description)
	}
}

func TestServiceAndEnvironmentFromRowRejectNullTimestamp(t *testing.T) {
	validTimestamp := pgtype.Timestamptz{
		Time:  time.Date(2026, time.August, 18, 0, 0, 0, 0, time.UTC),
		Valid: true,
	}

	t.Run("service created at", func(t *testing.T) {
		_, err := serviceFromRow(sqlcgen.Service{
			UpdatedAt: validTimestamp,
		})

		if err == nil || err.Error() != "map service: created_at is null" {
			t.Fatalf("serviceFromRow() error = %v, want null created_at error", err)
		}
	})

	t.Run("service updated at", func(t *testing.T) {
		_, err := serviceFromRow(sqlcgen.Service{
			CreatedAt: validTimestamp,
		})

		if err == nil || err.Error() != "map service: updated_at is null" {
			t.Fatalf("serviceFromRow() error = %v, want null updated_at error", err)
		}
	})

	t.Run("environment created at", func(t *testing.T) {
		_, err := environmentFromRow(sqlcgen.Environment{
			UpdatedAt: validTimestamp,
		})

		if err == nil || err.Error() != "map environment: created_at is null" {
			t.Fatalf("environmentFromRow() error = %v, want null created_at error", err)
		}
	})

	t.Run("environment updated at", func(t *testing.T) {
		_, err := environmentFromRow(sqlcgen.Environment{
			CreatedAt: validTimestamp,
		})

		if err == nil || err.Error() != "map environment: updated_at is null" {
			t.Fatalf("environmentFromRow() error = %v, want null updated_at error", err)
		}
	})
}

func TestEnvironmentFromRowMapsValues(t *testing.T) {
	createdAt := time.Date(2026, time.August, 18, 12, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)

	got, err := environmentFromRow(sqlcgen.Environment{
		ID:        11,
		ServiceID: 9,
		Slug:      "production",
		Name:      "Production",
		CreatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  updatedAt,
			Valid: true,
		},
	})
	if err != nil {
		t.Fatalf("environmentFromRow() error = %v", err)
	}

	if got.ID != 11 || got.ServiceID != 9 {
		t.Fatalf("environmentFromRow() IDs = (%d, %d), want (11, 9)", got.ID, got.ServiceID)
	}
	if got.Slug != "production" || got.Name != "Production" {
		t.Fatalf("environmentFromRow() = %#v, want production / Production", got)
	}
}

func TestServiceListLimitWithExtra(t *testing.T) {
	tests := []struct {
		name    string
		limit   int32
		want    int32
		wantErr error
	}{
		{name: "one", limit: 1, want: 2},
		{name: "regular limit", limit: 20, want: 21},
		{name: "zero", limit: 0, wantErr: catalog.ErrInvalidArgument},
		{name: "negative", limit: -1, wantErr: catalog.ErrInvalidArgument},
		{name: "maximum int32", limit: maxInt32, wantErr: catalog.ErrInvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := serviceListLimitWithExtra(tt.limit)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("serviceListLimitWithExtra() error = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("serviceListLimitWithExtra() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestServicePageFromRows(t *testing.T) {
	base := time.Date(2026, time.August, 18, 12, 0, 0, 0, time.UTC)
	rows := []sqlcgen.Service{
		testSQLCService(3, 7, "three", base.Add(2*time.Minute)),
		testSQLCService(2, 7, "two", base.Add(time.Minute)),
		testSQLCService(1, 7, "one", base),
	}

	page, err := servicePageFromRows(rows, 2)
	if err != nil {
		t.Fatalf("servicePageFromRows() error = %v", err)
	}

	if len(page.Services) != 2 {
		t.Fatalf("len(Services) = %d, want 2", len(page.Services))
	}
	if page.Services[0].ID != 3 || page.Services[1].ID != 2 {
		t.Fatalf(
			"Services IDs = [%d %d], want [3 2]",
			page.Services[0].ID,
			page.Services[1].ID,
		)
	}
	if page.Next == nil {
		t.Fatal("Next = nil, want cursor")
	}

	wantNext := catalog.ServiceCursor{
		CreatedAt: base.Add(time.Minute),
		ID:        2,
	}
	if *page.Next != wantNext {
		t.Fatalf("Next = %#v, want %#v", *page.Next, wantNext)
	}
}

func TestClassifyServiceError(t *testing.T) {
	unknownError := errors.New("connection lost")

	tests := []struct {
		name      string
		operation string
		err       error
		want      error
	}{
		{
			name:      "not found",
			operation: "get service by ID",
			err:       pgx.ErrNoRows,
			want:      catalog.ErrNotFound,
		},
		{
			name:      "unique violation",
			operation: "create service",
			err: &pgconn.PgError{
				Code: "23505",
			},
			want: catalog.ErrConflict,
		},
		{
			name:      "missing parent team",
			operation: "create service",
			err: &pgconn.PgError{
				Code:           "23503",
				ConstraintName: "services_team_id_fkey",
			},
			want: catalog.ErrNotFound,
		},
		{
			name:      "other foreign key violation",
			operation: "create environment",
			err: &pgconn.PgError{
				Code: "23503",
			},
			want: catalog.ErrConflict,
		},
		{
			name:      "unknown error",
			operation: "create service",
			err:       unknownError,
			want:      unknownError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyServiceError(tt.operation, tt.err)

			if !errors.Is(got, tt.want) {
				t.Fatalf(
					"classifyServiceError() = %v, want error matching %v",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestEnvironmentListLimitWithExtra(t *testing.T) {
	tests := []struct {
		name    string
		limit   int32
		want    int32
		wantErr error
	}{
		{name: "one", limit: 1, want: 2},
		{name: "regular limit", limit: 20, want: 21},
		{name: "zero", limit: 0, wantErr: catalog.ErrInvalidArgument},
		{name: "negative", limit: -1, wantErr: catalog.ErrInvalidArgument},
		{name: "maximum int32", limit: maxInt32, wantErr: catalog.ErrInvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := environmentListLimitWithExtra(tt.limit)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"environmentListLimitWithExtra() error = %v, want %v",
					err,
					tt.wantErr,
				)
			}
			if got != tt.want {
				t.Fatalf(
					"environmentListLimitWithExtra() = %d, want %d",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestEnvironmentPageFromRows(t *testing.T) {
	base := time.Date(
		2026,
		time.September,
		2,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	t.Run("with extra row", func(t *testing.T) {
		rows := []sqlcgen.Environment{
			testSQLCEnvironment(3, 7, "three", base.Add(2*time.Minute)),
			testSQLCEnvironment(2, 7, "two", base.Add(time.Minute)),
			testSQLCEnvironment(1, 7, "one", base),
		}

		page, err := environmentPageFromRows(rows, 2)
		if err != nil {
			t.Fatalf("environmentPageFromRows() error = %v", err)
		}
		if len(page.Environments) != 2 {
			t.Fatalf(
				"len(Environments) = %d, want 2",
				len(page.Environments),
			)
		}
		if page.Environments[0].ID != 3 ||
			page.Environments[1].ID != 2 {
			t.Fatalf(
				"Environments IDs = [%d %d], want [3 2]",
				page.Environments[0].ID,
				page.Environments[1].ID,
			)
		}

		wantNext := catalog.EnvironmentCursor{
			CreatedAt: base.Add(time.Minute),
			ID:        2,
		}
		if page.Next == nil {
			t.Fatal("Next = nil, want cursor")
		}
		if *page.Next != wantNext {
			t.Fatalf("Next = %#v, want %#v", *page.Next, wantNext)
		}
	})

	t.Run("without extra row", func(t *testing.T) {
		rows := []sqlcgen.Environment{
			testSQLCEnvironment(2, 7, "two", base.Add(time.Minute)),
			testSQLCEnvironment(1, 7, "one", base),
		}

		page, err := environmentPageFromRows(rows, 2)
		if err != nil {
			t.Fatalf("environmentPageFromRows() error = %v", err)
		}
		if len(page.Environments) != 2 {
			t.Fatalf(
				"len(Environments) = %d, want 2",
				len(page.Environments),
			)
		}
		if page.Next != nil {
			t.Fatalf("Next = %#v, want nil", page.Next)
		}
	})
}

func TestClassifyEnvironmentError(t *testing.T) {
	unknownError := errors.New("connection lost")

	tests := []struct {
		name      string
		operation string
		err       error
		want      error
	}{
		{
			name:      "not found",
			operation: "get environment by ID",
			err:       pgx.ErrNoRows,
			want:      catalog.ErrNotFound,
		},
		{
			name:      "unique violation",
			operation: "create environment",
			err: &pgconn.PgError{
				Code: "23505",
			},
			want: catalog.ErrConflict,
		},
		{
			name:      "missing parent service",
			operation: "create environment",
			err: &pgconn.PgError{
				Code:           "23503",
				ConstraintName: "environments_service_id_fkey",
			},
			want: catalog.ErrNotFound,
		},
		{
			name:      "other foreign key violation",
			operation: "delete environment",
			err: &pgconn.PgError{
				Code: "23503",
			},
			want: catalog.ErrConflict,
		},
		{
			name:      "unknown error",
			operation: "create environment",
			err:       unknownError,
			want:      unknownError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyEnvironmentError(tt.operation, tt.err)

			if !errors.Is(got, tt.want) {
				t.Fatalf(
					"classifyEnvironmentError() = %v, want error matching %v",
					got,
					tt.want,
				)
			}
		})
	}
}

func testSQLCTeam(id int64, slug string, createdAt time.Time) sqlcgen.Team {
	return sqlcgen.Team{
		ID:   id,
		Slug: slug,
		Name: slug,
		CreatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
	}
}

func testSQLCService(
	id int64,
	teamID int64,
	slug string,
	createdAt time.Time,
) sqlcgen.Service {
	return sqlcgen.Service{
		ID:     id,
		TeamID: teamID,
		Slug:   slug,
		Name:   slug,
		CreatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
	}
}

func testSQLCEnvironment(
	id int64,
	serviceID int64,
	slug string,
	createdAt time.Time,
) sqlcgen.Environment {
	return sqlcgen.Environment{
		ID:        id,
		ServiceID: serviceID,
		Slug:      slug,
		Name:      slug,
		CreatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
	}
}
