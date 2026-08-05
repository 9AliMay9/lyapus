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
