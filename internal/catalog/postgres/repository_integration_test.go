//go:build integration

package postgres

import (
	"context"
	"errors"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/9AliMay9/lyapus/internal/catalog"
	"github.com/jackc/pgx/v5/pgxpool"
)

const testDatabaseURLVariable = "LYAPUS_TEST_DATABASE_URL"

func TestTeamRepositoryIntegration(t *testing.T) {
	pool := openIntegrationPool(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resetCatalogTables(t, ctx, pool)

	repository := NewTeamRepository(pool)

	created, err := repository.CreateTeam(ctx, catalog.CreateTeamInput{
		Slug: "platform",
		Name: "Platform",
	})
	if err != nil {
		t.Fatalf("CreateTeam() error = %v", err)
	}

	if created.ID <= 0 {
		t.Fatalf("CreateTeam() ID = %d, want a positive ID", created.ID)
	}
	if created.CreatedAt.IsZero() {
		t.Fatal("CreateTeam() CreatedAt is zero")
	}
	if created.UpdatedAt.IsZero() {
		t.Fatal("CreateTeam() UpdatedAt is zero")
	}
	if created.UpdatedAt.Before(created.CreatedAt) {
		t.Fatalf(
			"CreateTeam() UpdatedAt = %s, before CreatedAt = %s",
			created.UpdatedAt,
			created.CreatedAt,
		)
	}

	got, err := repository.GetTeamByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetTeamByID() error = %v", err)
	}
	if got != created {
		t.Fatalf("GetTeamByID() = %#v, want %#v", got, created)
	}

	_, err = repository.CreateTeam(ctx, catalog.CreateTeamInput{
		Slug: "platform",
		Name: "Another Platform",
	})
	if !errors.Is(err, catalog.ErrConflict) {
		t.Fatalf("CreateTeam() duplicate error = %v, want ErrConflict", err)
	}

	_, err = repository.GetTeamByID(ctx, 999)
	if !errors.Is(err, catalog.ErrNotFound) {
		t.Fatalf("GetTeamByID() missing error = %v, want ErrNotFound", err)
	}
}

func openIntegrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv(testDatabaseURLVariable)
	if databaseURL == "" {
		t.Fatalf("%s must be set", testDatabaseURLVariable)
	}

	parsedURL, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatalf("parse %s: %v", testDatabaseURLVariable, err)
	}
	if parsedURL.Scheme != "postgres" && parsedURL.Scheme != "postgresql" {
		t.Fatalf("%s must use a PostgreSQL URL", testDatabaseURLVariable)
	}

	databaseName := strings.TrimPrefix(parsedURL.Path, "/")
	if !strings.HasSuffix(databaseName, "_test") {
		t.Fatalf("%s database must end with _test", testDatabaseURLVariable)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create integration pool: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping integration database: %v", err)
	}

	return pool
}

func resetCatalogTables(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	if _, err := pool.Exec(ctx, "TRUNCATE TABLE teams RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("reset catalog tables: %v", err)
	}
}
