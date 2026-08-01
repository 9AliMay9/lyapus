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
