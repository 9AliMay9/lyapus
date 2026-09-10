//go:build integration

package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestCatalogConstraintsIntegrationTeamSlug(t *testing.T) {
	pool := openIntegrationPool(t)

	tests := []struct {
		name          string
		slug          string
		wantViolation bool
	}{
		{name: "one character", slug: "a"},
		{name: "letters digits and hyphen", slug: "platform-9"},
		{name: "maximum length", slug: strings.Repeat("a", 63)},
		{name: "trailing hyphen", slug: "platform-"},
		{name: "empty", slug: "", wantViolation: true},
		{name: "too long", slug: strings.Repeat("a", 64), wantViolation: true},
		{name: "uppercase", slug: "Platform", wantViolation: true},
		{name: "leading digit", slug: "9platform", wantViolation: true},
		{name: "leading hyphen", slug: "-platform", wantViolation: true},
		{name: "underscore", slug: "platform_api", wantViolation: true},
		{name: "space", slug: "platform api", wantViolation: true},
		{name: "non ASCII", slug: "平台", wantViolation: true},
		{name: "trailing newline", slug: "platform\n", wantViolation: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resetCatalogTables(t, ctx, pool)

			_, err := pool.Exec(
				ctx,
				"INSERT INTO teams (slug, name) VALUES ($1, $2)",
				tt.slug,
				"Platform",
			)

			if !tt.wantViolation {
				if err != nil {
					t.Fatalf("insert valid slug %q: %v", tt.slug, err)
				}
				return
			}

			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) {
				t.Fatalf("insert slug %q error = %v, want PostgreSQL constraint error", tt.slug, err)
			}
			if pgErr.Code != "23514" {
				t.Fatalf("SQLSTATE = %q, want 23514", pgErr.Code)
			}
			if pgErr.ConstraintName != "teams_slug_format_check" {
				t.Fatalf(
					"constraint = %q, want teams_slug_format_check",
					pgErr.ConstraintName,
				)
			}
		})
	}
}

func TestCatalogConstraintsIntegrationTeamName(t *testing.T) {
	pool := openIntegrationPool(t)

	tests := []struct {
		label         string
		name          string
		wantViolation bool
	}{
		{label: "one character", name: "A"},
		{label: "internal space", name: "Platform Team"},
		{label: "maximum ASCII length", name: strings.Repeat("a", 100)},
		{label: "maximum Unicode length", name: strings.Repeat("团", 100)},
		{label: "empty", name: "", wantViolation: true},
		{label: "only spaces", name: "   ", wantViolation: true},
		{label: "leading space", name: " Platform", wantViolation: true},
		{label: "trailing space", name: "Platform ", wantViolation: true},
		{label: "too long ASCII", name: strings.Repeat("a", 101), wantViolation: true},
		{label: "too long Unicode", name: strings.Repeat("团", 101), wantViolation: true},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resetCatalogTables(t, ctx, pool)

			_, err := pool.Exec(
				ctx,
				"INSERT INTO teams (slug, name) VALUES ($1, $2)",
				"platform",
				tt.name,
			)

			if !tt.wantViolation {
				if err != nil {
					t.Fatalf("insert valid name %q: %v", tt.name, err)
				}
				return
			}

			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) {
				t.Fatalf(
					"insert name %q error = %v, want PostgreSQL constraint error",
					tt.name,
					err,
				)
			}
			if pgErr.Code != "23514" {
				t.Fatalf("SQLSTATE = %q, want 23514", pgErr.Code)
			}
			if pgErr.ConstraintName != "teams_name_check" {
				t.Fatalf(
					"constraint = %q, want teams_name_check",
					pgErr.ConstraintName,
				)
			}
		})
	}
}

func TestCatalogConstraintsIntegrationTeamTimestamps(t *testing.T) {
	pool := openIntegrationPool(t)
	createdAt := time.Date(2026, time.September, 9, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		offset        time.Duration
		wantViolation bool
	}{
		{name: "equal", offset: 0},
		{name: "later", offset: time.Microsecond},
		{name: "earlier", offset: -time.Microsecond, wantViolation: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resetCatalogTables(t, ctx, pool)

			_, err := pool.Exec(
				ctx,
				`INSERT INTO teams (slug, name, created_at, updated_at) VALUES ($1, $2, $3, $4)`,
				"platform",
				"Platform",
				createdAt,
				createdAt.Add(tt.offset),
			)

			if !tt.wantViolation {
				if err != nil {
					t.Fatalf("insert valid timestamps: %v", err)
				}
				return
			}

			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) {
				t.Fatalf(
					"insert invalid timestamps error = %v, want PostgreSQL constraint error",
					err,
				)
			}
			if pgErr.Code != "23514" {
				t.Fatalf("SQLSTATE = %q, want 23514", pgErr.Code)
			}
			if pgErr.ConstraintName != "teams_updated_at_check" {
				t.Fatalf(
					"constraint = %q, want teams_updated_at_check",
					pgErr.ConstraintName,
				)
			}
		})
	}
}

func TestCatalogConstraintsIntegrationServiceDescription(t *testing.T) {
	pool := openIntegrationPool(t)

	tests := []struct {
		name          string
		description   any
		wantViolation bool
	}{
		{name: "null", description: nil},
		{name: "empty", description: ""},
		{name: "maximum ASCII length", description: strings.Repeat("a", 500)},
		{name: "maximum Unicode length", description: strings.Repeat("述", 500)},
		{name: "too long ASCII", description: strings.Repeat("a", 501), wantViolation: true},
		{name: "too long Unicode", description: strings.Repeat("述", 501), wantViolation: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resetCatalogTables(t, ctx, pool)

			var teamID int64
			err := pool.QueryRow(
				ctx,
				"INSERT INTO teams (slug, name) VALUES ($1, $2) RETURNING id",
				"platform",
				"Platform",
			).Scan(&teamID)
			if err != nil {
				t.Fatalf("insert parent team: %v", err)
			}

			_, err = pool.Exec(
				ctx,
				`INSERT INTO services (team_id, slug, name, description) VALUES ($1, $2, $3, $4)`,
				teamID,
				"catalog",
				"Catalog",
				tt.description,
			)

			if !tt.wantViolation {
				if err != nil {
					t.Fatalf("insert valid description: %v", err)
				}
				return
			}

			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) {
				t.Fatalf(
					"insert invalid description error = %v, want PostgreSQL constraint error",
					err,
				)
			}
			if pgErr.Code != "23514" {
				t.Fatalf("SQLSTATE = %q, want 23514", pgErr.Code)
			}
			if pgErr.ConstraintName != "services_description_length_check" {
				t.Fatalf(
					"constraint = %q, want services_description_length_check",
					pgErr.ConstraintName,
				)
			}
		})
	}
}

func TestCatalogConstraintsIntegrationServiceSlug(t *testing.T) {
	pool := openIntegrationPool(t)

	tests := []struct {
		name          string
		slug          string
		wantViolation bool
	}{
		{name: "one character", slug: "a"},
		{name: "letters digits and hyphen", slug: "catalog-9"},
		{name: "maximum length", slug: strings.Repeat("a", 63)},
		{name: "trailing hyphen", slug: "catalog-"},
		{name: "empty", slug: "", wantViolation: true},
		{name: "too long", slug: strings.Repeat("a", 64), wantViolation: true},
		{name: "uppercase", slug: "Catalog", wantViolation: true},
		{name: "leading digit", slug: "9catalog", wantViolation: true},
		{name: "leading hyphen", slug: "-catalog", wantViolation: true},
		{name: "underscore", slug: "catalog_api", wantViolation: true},
		{name: "space", slug: "catalog api", wantViolation: true},
		{name: "non ASCII", slug: "目录", wantViolation: true},
		{name: "trailing newline", slug: "catalog\n", wantViolation: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resetCatalogTables(t, ctx, pool)

			var teamID int64
			err := pool.QueryRow(
				ctx,
				"INSERT INTO teams (slug, name) VALUES ($1, $2) RETURNING id",
				"platform",
				"Platform",
			).Scan(&teamID)
			if err != nil {
				t.Fatalf("insert parent team: %v", err)
			}

			_, err = pool.Exec(
				ctx,
				"INSERT INTO services (team_id, slug, name) VALUES ($1, $2, $3)",
				teamID,
				tt.slug,
				"Catalog",
			)

			if !tt.wantViolation {
				if err != nil {
					t.Fatalf("insert valid slug %q: %v", tt.slug, err)
				}
				return
			}

			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) {
				t.Fatalf(
					"insert slug %q error = %v, want PostgreSQL constraint error",
					tt.slug,
					err,
				)
			}
			if pgErr.Code != "23514" {
				t.Fatalf("SQLSTATE = %q, want 23514", pgErr.Code)
			}
			if pgErr.ConstraintName != "services_slug_format_check" {
				t.Fatalf(
					"constraint = %q, want services_slug_format_check",
					pgErr.ConstraintName,
				)
			}
		})
	}
}

func TestCatalogConstraintsIntegrationServiceName(t *testing.T) {
	pool := openIntegrationPool(t)

	tests := []struct {
		label         string
		name          string
		wantViolation bool
	}{
		{label: "one character", name: "A"},
		{label: "internal space", name: "Catalog API"},
		{label: "maximum ASCII length", name: strings.Repeat("a", 100)},
		{label: "maximum Unicode length", name: strings.Repeat("服", 100)},
		{label: "empty", name: "", wantViolation: true},
		{label: "only spaces", name: "   ", wantViolation: true},
		{label: "leading space", name: " Catalog", wantViolation: true},
		{label: "trailing space", name: "Catalog ", wantViolation: true},
		{label: "too long ASCII", name: strings.Repeat("a", 101), wantViolation: true},
		{label: "too long Unicode", name: strings.Repeat("服", 101), wantViolation: true},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resetCatalogTables(t, ctx, pool)

			var teamID int64
			err := pool.QueryRow(
				ctx,
				"INSERT INTO teams (slug, name) VALUES ($1, $2) RETURNING id",
				"platform",
				"Platform",
			).Scan(&teamID)
			if err != nil {
				t.Fatalf("insert parent team: %v", err)
			}

			_, err = pool.Exec(
				ctx,
				"INSERT INTO services (team_id, slug, name) VALUES ($1, $2, $3)",
				teamID,
				"catalog",
				tt.name,
			)

			if !tt.wantViolation {
				if err != nil {
					t.Fatalf("insert valid name %q: %v", tt.name, err)
				}
				return
			}

			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) {
				t.Fatalf(
					"insert name %q error %v, want PostgreSQL constraint error",
					tt.name,
					err,
				)
			}
			if pgErr.Code != "23514" {
				t.Fatalf("SQLSTATE = %q, want 23514", pgErr.Code)
			}
			if pgErr.ConstraintName != "services_name_check" {
				t.Fatalf(
					"constraint = %q, want services_name_check",
					pgErr.ConstraintName,
				)
			}
		})
	}
}

func TestCatalogConstraintsIntegrationServiceTimestamps(t *testing.T) {
	pool := openIntegrationPool(t)
	createdAt := time.Date(2026, time.September, 9, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		offset        time.Duration
		wantViolation bool
	}{
		{name: "equal", offset: 0},
		{name: "later", offset: time.Microsecond},
		{name: "earlier", offset: -time.Microsecond, wantViolation: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resetCatalogTables(t, ctx, pool)

			var teamID int64
			err := pool.QueryRow(
				ctx,
				"INSERT INTO teams (slug, name) VALUES ($1, $2) RETURNING id",
				"platform",
				"Platform",
			).Scan(&teamID)
			if err != nil {
				t.Fatalf("insert parent team: %v", err)
			}

			_, err = pool.Exec(
				ctx,
				`INSERT INTO services (team_id, slug, name, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`,
				teamID,
				"catalog",
				"Catalog",
				createdAt,
				createdAt.Add(tt.offset),
			)

			if !tt.wantViolation {
				if err != nil {
					t.Fatalf("insert valid timestamps: %v", err)
				}
				return
			}

			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) {
				t.Fatalf(
					"insert invalid timestamps error = %v, want PostgreSQL constraint error",
					err,
				)
			}
			if pgErr.Code != "23514" {
				t.Fatalf("SQLSTATE = %q, want 23514", pgErr.Code)
			}
			if pgErr.ConstraintName != "services_updated_at_check" {
				t.Fatalf(
					"constraint = %q, want services_updated_at_check",
					pgErr.ConstraintName,
				)
			}
		})
	}
}

func TestCatalogConstraintsIntegrationEnvironmentSlug(t *testing.T) {
	pool := openIntegrationPool(t)

	tests := []struct {
		name          string
		slug          string
		wantViolation bool
	}{
		{name: "one character", slug: "a"},
		{name: "letters digits and hyphen", slug: "staging-9"},
		{name: "maximum length", slug: strings.Repeat("a", 63)},
		{name: "trailing hyphen", slug: "staging-"},
		{name: "empty", slug: "", wantViolation: true},
		{name: "too long", slug: strings.Repeat("a", 64), wantViolation: true},
		{name: "uppercase", slug: "Staging", wantViolation: true},
		{name: "leading digit", slug: "9staging", wantViolation: true},
		{name: "leading hyphen", slug: "-staging", wantViolation: true},
		{name: "underscore", slug: "staging_env", wantViolation: true},
		{name: "space", slug: "staging env", wantViolation: true},
		{name: "non ASCII", slug: "环境", wantViolation: true},
		{name: "trailing newline", slug: "staging\n", wantViolation: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resetCatalogTables(t, ctx, pool)

			var teamID int64
			err := pool.QueryRow(
				ctx,
				"INSERT INTO teams (slug, name) VALUES ($1, $2) RETURNING id",
				"platform",
				"Platform",
			).Scan(&teamID)
			if err != nil {
				t.Fatalf("insert parent team: %v", err)
			}

			var serviceID int64
			err = pool.QueryRow(
				ctx,
				`INSERT INTO services (team_id, slug, name) VALUES ($1, $2, $3) RETURNING id`,
				teamID,
				"catalog",
				"Catalog",
			).Scan(&serviceID)
			if err != nil {
				t.Fatalf("insert parent service: %v", err)
			}

			_, err = pool.Exec(
				ctx,
				"INSERT INTO environments (service_id, slug, name) VALUES ($1, $2, $3)",
				serviceID,
				tt.slug,
				"Staging",
			)

			if !tt.wantViolation {
				if err != nil {
					t.Fatalf("insert valid slug %q: %v", tt.slug, err)
				}
				return
			}

			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) {
				t.Fatalf(
					"insert slug %q error = %v, want PostgreSQL constraint error",
					tt.slug,
					err,
				)
			}
			if pgErr.Code != "23514" {
				t.Fatalf("SQLSTATE = %q, want 23514", pgErr.Code)
			}
			if pgErr.ConstraintName != "environments_slug_format_check" {
				t.Fatalf(
					"constraint = %q, want environments_slug_format_check",
					pgErr.ConstraintName,
				)
			}
		})
	}
}

func TestCatalogConstraintsIntegrationEnvironmentName(t *testing.T) {
	pool := openIntegrationPool(t)

	tests := []struct {
		label         string
		name          string
		wantViolation bool
	}{
		{label: "one character", name: "A"},
		{label: "internal space", name: "Staging Environment"},
		{label: "maximum ASCII length", name: strings.Repeat("a", 100)},
		{label: "maximum Unicode length", name: strings.Repeat("环", 100)},
		{label: "empty", name: "", wantViolation: true},
		{label: "only spaces", name: "   ", wantViolation: true},
		{label: "leading space", name: " Staging", wantViolation: true},
		{label: "trailing space", name: "Staging ", wantViolation: true},
		{label: "too long ASCII", name: strings.Repeat("a", 101), wantViolation: true},
		{label: "too long Unicode", name: strings.Repeat("环", 101), wantViolation: true},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resetCatalogTables(t, ctx, pool)

			var teamID int64
			err := pool.QueryRow(
				ctx,
				"INSERT INTO teams (slug, name) VALUES ($1, $2) RETURNING id",
				"platform",
				"Platform",
			).Scan(&teamID)
			if err != nil {
				t.Fatalf("insert parent team: %v", err)
			}

			var serviceID int64
			err = pool.QueryRow(
				ctx,
				`INSERT INTO services (team_id, slug, name) VALUES ($1, $2, $3) RETURNING id`,
				teamID,
				"catalog",
				"Catalog",
			).Scan(&serviceID)
			if err != nil {
				t.Fatalf("insert parent service: %v", err)
			}

			_, err = pool.Exec(
				ctx,
				"INSERT INTO environments (service_id, slug, name) VALUES ($1, $2, $3)",
				serviceID,
				"staging",
				tt.name,
			)

			if !tt.wantViolation {
				if err != nil {
					t.Fatalf("insert valid name %q: %v", tt.name, err)
				}
				return
			}

			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) {
				t.Fatalf(
					"insert name %q error = %v, want PostgreSQL constraint error",
					tt.name,
					err,
				)
			}
			if pgErr.Code != "23514" {
				t.Fatalf("SQLSTATE = %q, want 23514", pgErr.Code)
			}
			if pgErr.ConstraintName != "environments_name_check" {
				t.Fatalf(
					"constraint = %q, want environments_name_check",
					pgErr.ConstraintName,
				)
			}
		})
	}
}

func TestCatalogConstraintsIntegrationEnvironmentTimestamps(t *testing.T) {
	pool := openIntegrationPool(t)
	createdAt := time.Date(2026, time.September, 9, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		offset        time.Duration
		wantViolation bool
	}{
		{name: "equal", offset: 0},
		{name: "later", offset: time.Microsecond},
		{name: "earlier", offset: -time.Microsecond, wantViolation: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resetCatalogTables(t, ctx, pool)

			var teamID int64
			err := pool.QueryRow(
				ctx,
				"INSERT INTO teams (slug, name) VALUES ($1, $2) RETURNING id",
				"platform",
				"Platform",
			).Scan(&teamID)
			if err != nil {
				t.Fatalf("insert parent team: %v", err)
			}

			var serviceID int64
			err = pool.QueryRow(
				ctx,
				`INSERT INTO services (team_id, slug, name) VALUES ($1, $2, $3) RETURNING id`,
				teamID,
				"catalog",
				"Catalog",
			).Scan(&serviceID)
			if err != nil {
				t.Fatalf("insert parent service: %v", err)
			}

			_, err = pool.Exec(
				ctx,
				`INSERT INTO environments (service_id, slug, name, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`,
				serviceID,
				"staging",
				"Staging",
				createdAt,
				createdAt.Add(tt.offset),
			)

			if !tt.wantViolation {
				if err != nil {
					t.Fatalf("insert valid timestamps: %v", err)
				}
				return
			}

			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) {
				t.Fatalf(
					"insert invalid timestamps error = %v, want PostgreSQL constraint error",
					err,
				)
			}
			if pgErr.Code != "23514" {
				t.Fatalf("SQLSTATE = %q, want 23514", pgErr.Code)
			}
			if pgErr.ConstraintName != "environments_updated_at_check" {
				t.Fatalf(
					"constraint = %q, want environments_updated_at_check",
					pgErr.ConstraintName,
				)
			}
		})
	}
}

func TestCatalogConstraintsIntegrationTeamUnique(t *testing.T) {
	pool := openIntegrationPool(t)

	tests := []struct {
		name          string
		secondSlug    string
		wantViolation bool
	}{
		{name: "different slug", secondSlug: "another"},
		{name: "duplicate slug", secondSlug: "platform", wantViolation: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resetCatalogTables(t, ctx, pool)

			_, err := pool.Exec(
				ctx,
				"INSERT INTO teams (slug, name) VALUES ($1, $2)",
				"platform",
				"Platform",
			)
			if err != nil {
				t.Fatalf("insert first team: %v", err)
			}

			_, err = pool.Exec(
				ctx,
				"INSERT INTO teams (slug, name) VALUES ($1, $2)",
				tt.secondSlug,
				"Platform",
			)

			if !tt.wantViolation {
				if err != nil {
					t.Fatalf("insert different slug: %v", err)
				}
				return
			}

			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) {
				t.Fatalf(
					"insert duplicate slug error = %v, want PostgreSQL constraint error",
					err,
				)
			}
			if pgErr.Code != "23505" {
				t.Fatalf("SQLSTATE = %q, want 23505", pgErr.Code)
			}
			if pgErr.ConstraintName != "teams_slug_key" {
				t.Fatalf(
					"constraint = %q, want teams_slug_key",
					pgErr.ConstraintName,
				)
			}
		})
	}
}

func TestCatalogConstraintsIntegrationServiceUnique(t *testing.T) {
	pool := openIntegrationPool(t)

	tests := []struct {
		name          string
		differentTeam bool
		secondSlug    string
		wantViolation bool
	}{
		{
			name:       "same team different slug",
			secondSlug: "another",
		},
		{
			name:          "same team duplicate slug",
			secondSlug:    "catalog",
			wantViolation: true,
		},
		{
			name:          "different team same slug",
			differentTeam: true,
			secondSlug:    "catalog",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resetCatalogTables(t, ctx, pool)

			var teamID int64
			err := pool.QueryRow(
				ctx,
				"INSERT INTO teams (slug, name) VALUES ($1, $2) RETURNING id",
				"platform",
				"Platform",
			).Scan(&teamID)
			if err != nil {
				t.Fatalf("insert first team: %v", err)
			}

			secondTeamID := teamID
			if tt.differentTeam {
				err = pool.QueryRow(
					ctx,
					"INSERT INTO teams (slug, name) VALUES ($1, $2) RETURNING id",
					"another",
					"Another",
				).Scan(&secondTeamID)
				if err != nil {
					t.Fatalf("insert second team: %v", err)
				}
			}

			_, err = pool.Exec(
				ctx,
				"INSERT INTO services (team_id, slug, name) VALUES ($1, $2, $3)",
				teamID,
				"catalog",
				"Catalog",
			)
			if err != nil {
				t.Fatalf("insert first service: %v", err)
			}

			_, err = pool.Exec(
				ctx,
				"INSERT INTO services (team_id, slug, name) VALUES ($1, $2, $3)",
				secondTeamID,
				tt.secondSlug,
				"Catalog",
			)

			if !tt.wantViolation {
				if err != nil {
					t.Fatalf("insert allowed service: %v", err)
				}
				return
			}

			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) {
				t.Fatalf(
					"insert duplicate service error = %v, want PostgreSQL constraint error",
					err,
				)
			}
			if pgErr.Code != "23505" {
				t.Fatalf("SQLSTATE = %q, want 23505", pgErr.Code)
			}
			if pgErr.ConstraintName != "services_team_id_slug_key" {
				t.Fatalf(
					"constraint = %q, want services_team_id_slug_key",
					pgErr.ConstraintName,
				)
			}
		})
	}
}

func TestCatalogConstraintsIntegrationEnvironmentUnique(t *testing.T) {
	pool := openIntegrationPool(t)

	tests := []struct {
		name             string
		differentService bool
		secondSlug       string
		wantViolation    bool
	}{
		{
			name:       "same service different slug",
			secondSlug: "production",
		},
		{
			name:          "same service duplicate slug",
			secondSlug:    "staging",
			wantViolation: true,
		},
		{
			name:             "different service same slug",
			differentService: true,
			secondSlug:       "staging",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resetCatalogTables(t, ctx, pool)

			var teamID int64
			err := pool.QueryRow(
				ctx,
				"INSERT INTO teams (slug, name) VALUES ($1, $2) RETURNING id",
				"platform",
				"Platform",
			).Scan(&teamID)
			if err != nil {
				t.Fatalf("insert parent team: %v", err)
			}

			var serviceID int64
			err = pool.QueryRow(
				ctx,
				`INSERT INTO services (team_id, slug, name) VALUES ($1, $2, $3) RETURNING id`,
				teamID,
				"catalog",
				"Catalog",
			).Scan(&serviceID)
			if err != nil {
				t.Fatalf("insert first service: %v", err)
			}

			secondServiceID := serviceID
			if tt.differentService {
				err = pool.QueryRow(
					ctx,
					`INSERT INTO services (team_id, slug, name) VALUES ($1, $2, $3) RETURNING id`,
					teamID,
					"another",
					"Another",
				).Scan(&secondServiceID)
				if err != nil {
					t.Fatalf("insert second service: %v", err)
				}
			}

			_, err = pool.Exec(
				ctx,
				"INSERT INTO environments (service_id, slug, name) VALUES ($1, $2, $3)",
				serviceID,
				"staging",
				"Staging",
			)
			if err != nil {
				t.Fatalf("insert first environment: %v", err)
			}

			_, err = pool.Exec(
				ctx,
				"INSERT INTO environments (service_id, slug, name) VALUES ($1, $2, $3)",
				secondServiceID,
				tt.secondSlug,
				"Staging",
			)

			if !tt.wantViolation {
				if err != nil {
					t.Fatalf("insert allowed environment: %v", err)
				}
				return
			}

			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) {
				t.Fatalf(
					"insert duplicate environment error = %v, want PostgreSQL constraint error",
					err,
				)
			}
			if pgErr.Code != "23505" {
				t.Fatalf("SQLSTATE = %q, want 23505", pgErr.Code)
			}
			if pgErr.ConstraintName != "environments_service_id_slug_key" {
				t.Fatalf(
					"constraint = %q, want environments_service_id_slug_key",
					pgErr.ConstraintName,
				)
			}
		})
	}
}

func TestCatalogConstraintsIntegrationServiceForeignKey(t *testing.T) {
	pool := openIntegrationPool(t)

	tests := []struct {
		name          string
		missingTeam   bool
		wantViolation bool
	}{
		{name: "existing team"},
		{
			name:          "missing team",
			missingTeam:   true,
			wantViolation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resetCatalogTables(t, ctx, pool)

			var teamID int64
			err := pool.QueryRow(
				ctx,
				"INSERT INTO teams (slug, name) VALUES ($1, $2) RETURNING id",
				"platform",
				"Platform",
			).Scan(&teamID)
			if err != nil {
				t.Fatalf("insert parent team: %v", err)
			}

			if tt.missingTeam {
				tag, err := pool.Exec(
					ctx,
					"DELETE FROM teams WHERE id = $1",
					teamID,
				)
				if err != nil {
					t.Fatalf("delete parent team: %v", err)
				}
				if tag.RowsAffected() != 1 {
					t.Fatalf(
						"deleted team count = %d, want 1",
						tag.RowsAffected(),
					)
				}
			}

			_, err = pool.Exec(
				ctx,
				"INSERT INTO services (team_id, slug, name) VALUES ($1, $2, $3)",
				teamID,
				"catalog",
				"Catalog",
			)

			if !tt.wantViolation {
				if err != nil {
					t.Fatalf("insert service with existing team: %v", err)
				}
				return
			}

			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) {
				t.Fatalf(
					"insert service with missing team error = %v, want PostgreSQL constraint error",
					err,
				)
			}
			if pgErr.Code != "23503" {
				t.Fatalf("SQLSTATE = %q, want 23503", pgErr.Code)
			}
			if pgErr.ConstraintName != "services_team_id_fkey" {
				t.Fatalf(
					"constraint = %q, want services_team_id_fkey",
					pgErr.ConstraintName,
				)
			}
		})
	}
}

func TestCatalogConstraintsIntegrationEnvironmentForeignKey(t *testing.T) {
	pool := openIntegrationPool(t)

	tests := []struct {
		name           string
		missingService bool
		wantViolation  bool
	}{
		{name: "existing service"},
		{
			name:           "missing service",
			missingService: true,
			wantViolation:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resetCatalogTables(t, ctx, pool)

			var teamID int64
			err := pool.QueryRow(
				ctx,
				"INSERT INTO teams (slug, name) VALUES ($1, $2) RETURNING id",
				"platform",
				"Platform",
			).Scan(&teamID)
			if err != nil {
				t.Fatalf("insert parent team: %v", err)
			}

			var serviceID int64
			err = pool.QueryRow(
				ctx,
				`INSERT INTO services (team_id, slug, name) VALUES ($1, $2, $3) RETURNING id`,
				teamID,
				"catalog",
				"Catalog",
			).Scan(&serviceID)
			if err != nil {
				t.Fatalf("insert parent service: %v", err)
			}

			if tt.missingService {
				tag, err := pool.Exec(
					ctx,
					"DELETE FROM services WHERE id = $1",
					serviceID,
				)
				if err != nil {
					t.Fatalf("delete parent service: %v", err)
				}
				if tag.RowsAffected() != 1 {
					t.Fatalf(
						"deleted service count = %d, want 1",
						tag.RowsAffected(),
					)
				}
			}

			_, err = pool.Exec(
				ctx,
				"INSERT INTO environments (service_id, slug, name) VALUES ($1, $2, $3)",
				serviceID,
				"staging",
				"Staging",
			)

			if !tt.wantViolation {
				if err != nil {
					t.Fatalf("insert environment with existing service: %v", err)
				}
				return
			}

			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) {
				t.Fatalf(
					"insert environment with missing service error = %v, want PostgreSQL constraint error",
					err,
				)
			}
			if pgErr.Code != "23503" {
				t.Fatalf("SQLSTATE = %q, want 23503", pgErr.Code)
			}
			if pgErr.ConstraintName != "environments_service_id_fkey" {
				t.Fatalf(
					"constraint = %q, want environments_service_id_fkey",
					pgErr.ConstraintName,
				)
			}
		})
	}
}

func TestCatalogConstraintsIntegrationTeamDeleteRestriction(t *testing.T) {
	pool := openIntegrationPool(t)

	tests := []struct {
		name       string
		hasService bool
	}{
		{name: "without service"},
		{name: "with service", hasService: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resetCatalogTables(t, ctx, pool)

			var teamID int64
			err := pool.QueryRow(
				ctx,
				"INSERT INTO teams (slug, name) VALUES ($1, $2) RETURNING id",
				"platform",
				"Platform",
			).Scan(&teamID)
			if err != nil {
				t.Fatalf("insert parent team: %v", err)
			}

			if tt.hasService {
				_, err = pool.Exec(
					ctx,
					"INSERT INTO services (team_id, slug, name) VALUES ($1, $2, $3)",
					teamID,
					"catalog",
					"Catalog",
				)
				if err != nil {
					t.Fatalf("insert child service: %v", err)
				}
			}

			tag, err := pool.Exec(
				ctx,
				"DELETE FROM teams WHERE id = $1",
				teamID,
			)

			if !tt.hasService {
				if err != nil {
					t.Fatalf("delete unreferenced team: %v", err)
				}
				if tag.RowsAffected() != 1 {
					t.Fatalf(
						"deleted team count = %d, want 1",
						tag.RowsAffected(),
					)
				}
			} else {
				var pgErr *pgconn.PgError
				if !errors.As(err, &pgErr) {
					t.Fatalf(
						"delete referenced team error = %v, want PostgreSQL constraint error",
						err,
					)
				}
				if pgErr.Code != "23503" {
					t.Fatalf("SQLSTATE = %q, want 23503", pgErr.Code)
				}
				if pgErr.ConstraintName != "services_team_id_fkey" {
					t.Fatalf(
						"constraint = %q, want services_team_id_fkey",
						pgErr.ConstraintName,
					)
				}
			}

			var teamCount, serviceCount int64
			err = pool.QueryRow(
				ctx,
				`SELECT (SELECT count(*) FROM teams WHERE id = $1), (SELECT count(*) FROM services WHERE team_id = $1)`,
				teamID,
			).Scan(&teamCount, &serviceCount)
			if err != nil {
				t.Fatalf("read remaining rows: %v", err)
			}

			var wantCount int64
			if tt.hasService {
				wantCount = 1
			}
			if teamCount != wantCount || serviceCount != wantCount {
				t.Fatalf(
					"remaining teams = %d, services = %d, want both %d",
					teamCount,
					serviceCount,
					wantCount,
				)
			}
		})
	}
}

func TestCatalogConstraintsIntegrationServiceDeleteRestriction(t *testing.T) {
	pool := openIntegrationPool(t)

	tests := []struct {
		name           string
		hasEnvironment bool
	}{
		{name: "without environment"},
		{name: "with environment", hasEnvironment: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resetCatalogTables(t, ctx, pool)

			var teamID int64
			err := pool.QueryRow(
				ctx,
				"INSERT INTO teams (slug, name) VALUES ($1, $2) RETURNING id",
				"platform",
				"Platform",
			).Scan(&teamID)
			if err != nil {
				t.Fatalf("insert parent team: %v", err)
			}

			var serviceID int64
			err = pool.QueryRow(
				ctx,
				`INSERT INTO services (team_id, slug, name) VALUES ($1, $2, $3) RETURNING id`,
				teamID,
				"catalog",
				"Catalog",
			).Scan(&serviceID)
			if err != nil {
				t.Fatalf("insert parent service: %v", err)
			}

			if tt.hasEnvironment {
				_, err = pool.Exec(
					ctx,
					"INSERT INTO environments (service_id, slug, name) VALUES ($1, $2, $3)",
					serviceID,
					"staging",
					"Staging",
				)
				if err != nil {
					t.Fatalf("insert child environment: %v", err)
				}
			}

			tag, err := pool.Exec(
				ctx,
				"DELETE FROM services WHERE id = $1",
				serviceID,
			)

			if !tt.hasEnvironment {
				if err != nil {
					t.Fatalf("delete unreferenced service: %v", err)
				}
				if tag.RowsAffected() != 1 {
					t.Fatalf(
						"deleted service count = %d, want 1",
						tag.RowsAffected(),
					)
				}
			} else {
				var pgErr *pgconn.PgError
				if !errors.As(err, &pgErr) {
					t.Fatalf(
						"delete referenced service error = %v, want PostgreSQL constraint error",
						err,
					)
				}
				if pgErr.Code != "23503" {
					t.Fatalf("SQLSTATE = %q, want 23503", pgErr.Code)
				}
				if pgErr.ConstraintName != "environments_service_id_fkey" {
					t.Fatalf(
						"constraint = %q, want environments_service_id_fkey",
						pgErr.ConstraintName,
					)
				}
			}

			var serviceCount, environmentCount int64
			err = pool.QueryRow(
				ctx,
				`SELECT (SELECT count(*) FROM services WHERE id = $1), (SELECT count(*) FROM environments WHERE service_id = $1)`,
				serviceID,
			).Scan(&serviceCount, &environmentCount)
			if err != nil {
				t.Fatalf("read remaining rows: %v", err)
			}

			var wantCount int64
			if tt.hasEnvironment {
				wantCount = 1
			}
			if serviceCount != wantCount || environmentCount != wantCount {
				t.Fatalf(
					"remaining services = %d, environments = %d, want both %d",
					serviceCount,
					environmentCount,
					wantCount,
				)
			}
		})
	}
}
