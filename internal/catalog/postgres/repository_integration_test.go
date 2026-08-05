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

func TestTeamRepositoryIntegrationCreateGetAndErrors(t *testing.T) {
	repository, _, ctx := setupTeamRepositoryIntegration(t)

	created := createIntegrationTeam(t, ctx, repository, "platform", "Platform")

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

func TestTeamRepositoryIntegrationListPagination(t *testing.T) {
	repository, _, ctx := setupTeamRepositoryIntegration(t)

	created := []catalog.Team{
		createIntegrationTeam(t, ctx, repository, "one", "One"),
		createIntegrationTeam(t, ctx, repository, "two", "Two"),
		createIntegrationTeam(t, ctx, repository, "three", "Three"),
	}

	first, err := repository.ListTeams(ctx, catalog.ListTeamsInput{Limit: 2})
	if err != nil {
		t.Fatalf("ListTeams() first page error = %v", err)
	}
	if len(first.Teams) != 2 {
		t.Fatalf("ListTeams() first page length = %d, want 2", len(first.Teams))
	}
	if first.Next == nil {
		t.Fatal("ListTeams() first page Next = nil, want cursor")
	}
	assertTeamsInDescendingCursorOrder(t, first.Teams)

	second, err := repository.ListTeams(ctx, catalog.ListTeamsInput{
		Limit: 2,
		After: first.Next,
	})
	if err != nil {
		t.Fatalf("ListTeams() second page error = %v", err)
	}
	if len(second.Teams) != 1 {
		t.Fatalf("ListTeams() second page length = %d, want 1", len(second.Teams))
	}
	if second.Next != nil {
		t.Fatalf("ListTeams() second page Next = %#v, want nil", second.Next)
	}

	all := append(append([]catalog.Team{}, first.Teams...), second.Teams...)
	if len(all) != len(created) {
		t.Fatalf("ListTeams() total length = %d, want %d", len(all), len(created))
	}

	wantIDs := map[int64]struct{}{}
	for _, team := range created {
		wantIDs[team.ID] = struct{}{}
	}
	for _, team := range all {
		if _, ok := wantIDs[team.ID]; !ok {
			t.Fatalf("ListTeams() returned unexpected team ID %d", team.ID)
		}
		delete(wantIDs, team.ID)
	}
	if len(wantIDs) != 0 {
		t.Fatalf("ListTeams() missed team IDs %#v", wantIDs)
	}

	_, err = repository.ListTeams(ctx, catalog.ListTeamsInput{Limit: 0})
	if !errors.Is(err, catalog.ErrInvalidArgument) {
		t.Fatalf("ListTeams() invalid limit error = %v, want ErrInvalidArgument", err)
	}
}

func TestTeamRepositoryIntegrationUpdateAndDelete(t *testing.T) {
	repository, pool, ctx := setupTeamRepositoryIntegration(t)

	created := createIntegrationTeam(t, ctx, repository, "platform", "Platform")

	updated, err := repository.UpdateTeam(ctx, created.ID, catalog.UpdateTeamInput{
		Slug: "platform-engineering",
		Name: "Platform Engineering",
	})
	if err != nil {
		t.Fatalf("UpdateTeam() error = %v", err)
	}
	if updated.ID != created.ID {
		t.Fatalf("UpdateTeam() ID = %d, want %d", updated.ID, created.ID)
	}
	if updated.Slug != "platform-engineering" {
		t.Fatalf("UpdateTeam() Slug = %q, want %q", updated.Slug, "platform-engineering")
	}
	if updated.Name != "Platform Engineering" {
		t.Fatalf("UpdateTeam() Name = %q, want %q", updated.Name, "Platform Engineering")
	}
	if updated.CreatedAt != created.CreatedAt {
		t.Fatalf("UpdateTeam() CreatedAt = %s, want %s", updated.CreatedAt, created.CreatedAt)
	}
	if updated.UpdatedAt.Before(created.UpdatedAt) {
		t.Fatalf("UpdateTeam() UpdatedAt = %s, before previous value %s", updated.UpdatedAt, created.UpdatedAt)
	}

	got, err := repository.GetTeamByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetTeamByID() after update error = %v", err)
	}
	if got != updated {
		t.Fatalf("GetTeamByID() after update = %#v, want %#v", got, updated)
	}

	other := createIntegrationTeam(t, ctx, repository, "other", "Other")
	_, err = repository.UpdateTeam(ctx, other.ID, catalog.UpdateTeamInput{
		Slug: updated.Slug,
		Name: "Other",
	})
	if !errors.Is(err, catalog.ErrConflict) {
		t.Fatalf("UpdateTeam() duplicate slug error = %v, want ErrConflict", err)
	}

	_, err = repository.UpdateTeam(ctx, 999, catalog.UpdateTeamInput{
		Slug: "missing",
		Name: "Missing",
	})
	if !errors.Is(err, catalog.ErrNotFound) {
		t.Fatalf("UpdateTeam() missing error = %v, want ErrNotFound", err)
	}

	if _, err := pool.Exec(
		ctx,
		"INSERT INTO services (team_id, slug, name) VALUES ($1, $2, $3)",
		created.ID,
		"catalog-api",
		"Catalog API",
	); err != nil {
		t.Fatalf("seed service for foreign key test: %v", err)
	}

	err = repository.DeleteTeam(ctx, created.ID)
	if !errors.Is(err, catalog.ErrConflict) {
		t.Fatalf("DeleteTeam() referenced error = %v, want ErrConflict", err)
	}

	deletable := createIntegrationTeam(t, ctx, repository, "deletable", "Deletable")
	if err := repository.DeleteTeam(ctx, deletable.ID); err != nil {
		t.Fatalf("DeleteTeam() error = %v", err)
	}

	_, err = repository.GetTeamByID(ctx, deletable.ID)
	if !errors.Is(err, catalog.ErrNotFound) {
		t.Fatalf("GetTeamByID() after delete error = %v, want ErrNotFound", err)
	}

	err = repository.DeleteTeam(ctx, deletable.ID)
	if !errors.Is(err, catalog.ErrNotFound) {
		t.Fatalf("DeleteTeam() missing error = %v, want ErrNotFound", err)
	}
}

func setupTeamRepositoryIntegration(t *testing.T) (*TeamRepository, *pgxpool.Pool, context.Context) {
	t.Helper()

	pool := openIntegrationPool(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	resetCatalogTables(t, ctx, pool)

	return NewTeamRepository(pool), pool, ctx
}

func createIntegrationTeam(
	t *testing.T,
	ctx context.Context,
	repository *TeamRepository,
	slug string,
	name string,
) catalog.Team {
	t.Helper()

	team, err := repository.CreateTeam(ctx, catalog.CreateTeamInput{
		Slug: slug,
		Name: name,
	})
	if err != nil {
		t.Fatalf("CreateTeam(%q) error = %v", slug, err)
	}

	return team
}

func assertTeamsInDescendingCursorOrder(t *testing.T, teams []catalog.Team) {
	t.Helper()

	for index := 0; index+1 < len(teams); index++ {
		current := teams[index]
		next := teams[index+1]

		if current.CreatedAt.Before(next.CreatedAt) {
			t.Fatalf(
				"Teams are not ordered by created_at DESC: %s before %s",
				current.CreatedAt,
				next.CreatedAt,
			)
		}
		if current.CreatedAt.Equal(next.CreatedAt) && current.ID <= next.ID {
			t.Fatalf(
				"Teams with equal created_at are not ordered by id DESC: %d before %d",
				current.ID,
				next.ID,
			)
		}
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
