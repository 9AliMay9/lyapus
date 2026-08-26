//go:build integration

package postgres

import (
	"context"
	"errors"
	"net/url"
	"os"
	"strings"
	"sync"
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

	updatedName := "Platform Engineering"
	updated, err := repository.UpdateTeam(ctx, created.ID, catalog.UpdateTeamInput{
		Name: &updatedName,
	})
	if err != nil {
		t.Fatalf("UpdateTeam() error = %v", err)
	}
	if updated.ID != created.ID {
		t.Fatalf("UpdateTeam() ID = %d, want %d", updated.ID, created.ID)
	}
	if updated.Slug != created.Slug {
		t.Fatalf("UpdateTeam() Slug = %q, want %q", updated.Slug, created.Slug)
	}
	if updated.Name != updatedName {
		t.Fatalf("UpdateTeam() Name = %q, want %q", updated.Name, updatedName)
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
	duplicateSlug := updated.Slug
	otherName := "Other"
	_, err = repository.UpdateTeam(ctx, other.ID, catalog.UpdateTeamInput{
		Slug: &duplicateSlug,
		Name: &otherName,
	})
	if !errors.Is(err, catalog.ErrConflict) {
		t.Fatalf("UpdateTeam() duplicate slug error = %v, want ErrConflict", err)
	}

	missingSlug := "missing"
	missingName := "Missing"
	_, err = repository.UpdateTeam(ctx, 999, catalog.UpdateTeamInput{
		Slug: &missingSlug,
		Name: &missingName,
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

func TestServiceRepositoryIntegrationCreateGetAndTransaction(t *testing.T) {
	serviceRepository, teamRepository, pool, ctx := setupServiceRepositoryIntegration(t)

	team := createIntegrationTeam(
		t,
		ctx,
		teamRepository,
		"platform",
		"Platform",
	)

	description := "Service catalog API"
	created, err := serviceRepository.CreateService(
		ctx,
		catalog.CreateServiceInput{
			TeamID:      team.ID,
			Slug:        "catalog-api",
			Name:        "Catalog API",
			Description: &description,
			Environments: []catalog.CreateEnvironmentInput{
				{
					Slug: "staging",
					Name: "Staging",
				},
				{
					Slug: "production",
					Name: "Production",
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("CreateService() error = %v", err)
	}

	if created.Service.ID <= 0 {
		t.Fatalf("CreateService() ID = %d, want a positive ID", created.Service.ID)
	}
	if created.Service.TeamID != team.ID {
		t.Fatalf(
			"CreateService() TeamID = %d, want %d",
			created.Service.TeamID,
			team.ID,
		)
	}
	if created.Service.Description == nil || *created.Service.Description != description {
		t.Fatalf(
			"CreateService() Description = %#v, want %q",
			created.Service.Description,
			description,
		)
	}
	if created.Service.CreatedAt.IsZero() || created.Service.UpdatedAt.IsZero() {
		t.Fatalf("CreateService() timestamps = %#v, want non-zero values", created.Service)
	}
	if created.Service.UpdatedAt.Before(created.Service.CreatedAt) {
		t.Fatalf(
			"CreateService() UpdatedAt = %s, before CreatedAt = %s",
			created.Service.UpdatedAt,
			created.Service.CreatedAt,
		)
	}
	if len(created.Environments) != 2 {
		t.Fatalf(
			"CreateService() environment count = %d, want 2",
			len(created.Environments),
		)
	}
	for _, environment := range created.Environments {
		if environment.ID <= 0 {
			t.Fatalf("CreateService() environment ID = %d, want positive", environment.ID)
		}
		if environment.ServiceID != created.Service.ID {
			t.Fatalf(
				"CreateService() environment ServiceID = %d, want %d",
				environment.ServiceID,
				created.Service.ID,
			)
		}
	}

	got, err := serviceRepository.GetServiceByID(ctx, created.Service.ID)
	if err != nil {
		t.Fatalf("GetServiceByID() error = %v", err)
	}
	if got.Service.ID != created.Service.ID {
		t.Fatalf(
			"GetServiceByID() Service ID = %d, want %d",
			got.Service.ID,
			created.Service.ID,
		)
	}
	if got.Service.Description == nil || *got.Service.Description != description {
		t.Fatalf(
			"GetServiceByID() Description = %#v, want %q",
			got.Service.Description,
			description,
		)
	}
	if len(got.Environments) != 2 {
		t.Fatalf(
			"GetServiceByID() environment count = %d, want 2",
			len(got.Environments),
		)
	}
	if got.Environments[0].Slug != "staging" || got.Environments[1].Slug != "production" {
		t.Fatalf(
			"GetServiceByID() environment order = %#v, want staging then production",
			got.Environments,
		)
	}

	_, err = serviceRepository.CreateService(
		ctx,
		catalog.CreateServiceInput{
			TeamID: 999,
			Slug:   "missing-parent",
			Name:   "Missing Parent",
		},
	)
	if !errors.Is(err, catalog.ErrNotFound) {
		t.Fatalf(
			"CreateService() missing parent error = %v, want ErrNotFound",
			err,
		)
	}

	_, err = serviceRepository.CreateService(
		ctx,
		catalog.CreateServiceInput{
			TeamID: team.ID,
			Slug:   "must-roll-back",
			Name:   "Must Roll Back",
			Environments: []catalog.CreateEnvironmentInput{
				{
					Slug: "staging",
					Name: "Staging",
				},
				{
					Slug: "staging",
					Name: "Another Staging",
				},
			},
		},
	)
	if !errors.Is(err, catalog.ErrConflict) {
		t.Fatalf(
			"CreateService() duplicate initial environment error = %v, want ErrConflict",
			err,
		)
	}

	var count int
	if err := pool.QueryRow(
		ctx,
		"SELECT count(*) FROM services WHERE team_id = $1 AND slug = $2",
		team.ID,
		"must-roll-back",
	).Scan(&count); err != nil {
		t.Fatalf("count rolled back services: %v", err)
	}
	if count != 0 {
		t.Fatalf("rolled back service count = %d, want 0", count)
	}

	_, err = serviceRepository.GetServiceByID(ctx, 999)
	if !errors.Is(err, catalog.ErrNotFound) {
		t.Fatalf("GetServiceByID() missing error = %v, want ErrNotFound", err)
	}
}

func TestServiceRepositoryIntegrationUpdateAndDelete(t *testing.T) {
	serviceRepository, teamRepository, pool, ctx := setupServiceRepositoryIntegration(t)

	team := createIntegrationTeam(t, ctx, teamRepository, "platform", "Platform")

	description := "Original description"
	createdDetail, err := serviceRepository.CreateService(
		ctx,
		catalog.CreateServiceInput{
			TeamID:      team.ID,
			Slug:        "catalog-api",
			Name:        "Catalog API",
			Description: &description,
		},
	)
	if err != nil {
		t.Fatalf("CreateService() error = %v", err)
	}
	created := createdDetail.Service

	updatedName := "Catalog API v2"
	updated, err := serviceRepository.UpdateService(
		ctx,
		created.ID,
		catalog.UpdateServiceInput{
			Name: &updatedName,
		},
	)
	if err != nil {
		t.Fatalf("UpdateService() name error = %v", err)
	}
	if updated.ID != created.ID {
		t.Fatalf("UpdateService() ID = %d, want %d", updated.ID, created.ID)
	}
	if updated.TeamID != created.TeamID {
		t.Fatalf("UpdateService() TeamID = %d, want %d", updated.TeamID, created.TeamID)
	}
	if updated.Slug != created.Slug {
		t.Fatalf("UpdateService() Slug = %q, want %q", updated.Slug, created.Slug)
	}
	if updated.Name != updatedName {
		t.Fatalf("UpdateService() Name = %q, want %q", updated.Name, updatedName)
	}
	if updated.Description == nil || *updated.Description != description {
		t.Fatalf(
			"UpdateService() Description = %#v, want %q",
			updated.Description,
			description,
		)
	}
	if updated.CreatedAt != created.CreatedAt {
		t.Fatalf(
			"UpdateService() CreatedAt = %s, want %s",
			updated.CreatedAt,
			created.CreatedAt,
		)
	}
	if updated.UpdatedAt.Before(created.UpdatedAt) {
		t.Fatalf(
			"UpdateService() UpdatedAt = %s, before previous value %s",
			updated.UpdatedAt,
			created.UpdatedAt,
		)
	}

	cleared, err := serviceRepository.UpdateService(
		ctx,
		created.ID,
		catalog.UpdateServiceInput{
			DescriptionProvided: true,
		},
	)
	if err != nil {
		t.Fatalf("UpdateService() clear description error = %v", err)
	}
	if cleared.Description != nil {
		t.Fatalf("UpdateService() cleared Description = %#v, want nil", cleared.Description)
	}
	if cleared.Name != updatedName {
		t.Fatalf("UpdateService() cleared Name = %q, want %q", cleared.Name, updatedName)
	}

	got, err := serviceRepository.GetServiceByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetServiceByID() after update error = %v", err)
	}
	if got.Service != cleared {
		t.Fatalf(
			"GetServiceByID() after update = %#v, want %#v",
			got.Service,
			cleared,
		)
	}

	other := createIntegrationService(
		t,
		ctx,
		serviceRepository,
		team.ID,
		"other-api",
		"Other API",
	)
	duplicateSlug := cleared.Slug
	_, err = serviceRepository.UpdateService(
		ctx,
		other.ID,
		catalog.UpdateServiceInput{
			Slug: &duplicateSlug,
		},
	)
	if !errors.Is(err, catalog.ErrConflict) {
		t.Fatalf("UpdateService() duplicate slug error = %v, want ErrConflict", err)
	}

	missingName := "Missing"
	_, err = serviceRepository.UpdateService(
		ctx,
		999,
		catalog.UpdateServiceInput{
			Name: &missingName,
		},
	)
	if !errors.Is(err, catalog.ErrNotFound) {
		t.Fatalf("UpdateService() missing error = %#v, want ErrNotFound", err)
	}

	if _, err := pool.Exec(
		ctx,
		"INSERT INTO environments (service_id, slug, name) VALUES ($1, $2, $3)",
		created.ID,
		"production",
		"Production",
	); err != nil {
		t.Fatalf("seed environment for foreign key test: %v", err)
	}

	err = serviceRepository.DeleteService(ctx, created.ID)
	if !errors.Is(err, catalog.ErrConflict) {
		t.Fatalf("DeleteService() referenced error = %v, want ErrConflict", err)
	}

	deletable := createIntegrationService(
		t,
		ctx,
		serviceRepository,
		team.ID,
		"deletable-api",
		"Deletable API",
	)
	if err := serviceRepository.DeleteService(ctx, deletable.ID); err != nil {
		t.Fatalf("DeleteService() error = %v", err)
	}

	_, err = serviceRepository.GetServiceByID(ctx, deletable.ID)
	if !errors.Is(err, catalog.ErrNotFound) {
		t.Fatalf("GetServiceByID() after delete error = %v, want ErrNotFound", err)
	}

	err = serviceRepository.DeleteService(ctx, deletable.ID)
	if !errors.Is(err, catalog.ErrNotFound) {
		t.Fatalf("DeleteService() missing error = %v, want ErrNotFound", err)
	}
}

func TestServiceRepositoryIntegrationListPaginationAndTeamFilter(t *testing.T) {
	serviceRepository, teamRepository, _, ctx := setupServiceRepositoryIntegration(t)

	platform := createIntegrationTeam(t, ctx, teamRepository, "platform", "Platform")
	observability := createIntegrationTeam(
		t,
		ctx,
		teamRepository,
		"observability",
		"Observability",
	)

	created := []catalog.Service{
		createIntegrationService(
			t,
			ctx,
			serviceRepository,
			platform.ID,
			"catalog-api",
			"Catalog API",
		),
		createIntegrationService(
			t,
			ctx,
			serviceRepository,
			platform.ID,
			"deployment-api",
			"Deployment API",
		),
		createIntegrationService(
			t,
			ctx,
			serviceRepository,
			platform.ID,
			"incident-api",
			"Incident API",
		),
		createIntegrationService(
			t,
			ctx,
			serviceRepository,
			observability.ID,
			"metrics-api",
			"Metrics API",
		),
	}

	first, err := serviceRepository.ListServices(
		ctx,
		catalog.ListServicesInput{Limit: 2},
	)
	if err != nil {
		t.Fatalf("ListServices() global first page error = %v", err)
	}
	if len(first.Services) != 2 {
		t.Fatalf(
			"ListServices() global first page length = %d, want 2",
			len(first.Services),
		)
	}
	if first.Next == nil {
		t.Fatal("ListServices() global first page Next = nil, want cursor")
	}
	assertServicesInDescendingCursorOrder(t, first.Services)

	second, err := serviceRepository.ListServices(
		ctx,
		catalog.ListServicesInput{
			Limit: 2,
			After: first.Next,
		},
	)
	if err != nil {
		t.Fatalf("ListServices() global second page error = %v", err)
	}
	if len(second.Services) != 2 {
		t.Fatalf(
			"ListServices() global second page length = %d, want 2",
			len(second.Services),
		)
	}
	if second.Next != nil {
		t.Fatalf(
			"ListServices() global second page Next = %#v, want nil",
			second.Next,
		)
	}
	all := append(
		append([]catalog.Service{}, first.Services...),
		second.Services...,
	)
	if len(all) != len(created) {
		t.Fatalf(
			"ListServices() global total length = %d, want %d",
			len(all),
			len(created),
		)
	}

	wantIDs := map[int64]struct{}{}
	for _, service := range created {
		wantIDs[service.ID] = struct{}{}
	}
	for _, service := range all {
		if _, ok := wantIDs[service.ID]; !ok {
			t.Fatalf(
				"ListServices() global returned unexpected service ID %d",
				service.ID,
			)
		}
		delete(wantIDs, service.ID)
	}
	if len(wantIDs) != 0 {
		t.Fatalf("ListServices() global missed service IDs %#v", wantIDs)
	}

	teamID := platform.ID
	filteredFirst, err := serviceRepository.ListServices(
		ctx,
		catalog.ListServicesInput{
			TeamID: &teamID,
			Limit:  2,
		},
	)
	if err != nil {
		t.Fatalf("ListServices() filtered first page error = %v", err)
	}
	if len(filteredFirst.Services) != 2 {
		t.Fatalf(
			"ListServices() filtered first page length = %d, want 2",
			len(filteredFirst.Services),
		)
	}
	if filteredFirst.Next == nil {
		t.Fatal("ListServices() filtered first page Next = nil, want cursor")
	}
	assertServicesInDescendingCursorOrder(t, filteredFirst.Services)

	filteredSecond, err := serviceRepository.ListServices(
		ctx,
		catalog.ListServicesInput{
			TeamID: &teamID,
			Limit:  2,
			After:  filteredFirst.Next,
		},
	)
	if err != nil {
		t.Fatalf("ListServices() filtered second page error = %v", err)
	}
	if len(filteredSecond.Services) != 1 {
		t.Fatalf(
			"ListServices() filtered second page length = %d, want 1",
			len(filteredSecond.Services),
		)
	}
	if filteredSecond.Next != nil {
		t.Fatalf(
			"ListServices() filtered second page Next = %#v, want nil",
			filteredSecond.Next,
		)
	}

	filtered := append(
		append([]catalog.Service{}, filteredFirst.Services...),
		filteredSecond.Services...,
	)

	wantPlatformIDs := map[int64]struct{}{
		created[0].ID: {},
		created[1].ID: {},
		created[2].ID: {},
	}

	for _, service := range filtered {
		if service.TeamID != platform.ID {
			t.Fatalf(
				"ListServices() filtered TeamID = %d, want %d",
				service.TeamID,
				platform.ID,
			)
		}
		if _, ok := wantPlatformIDs[service.ID]; !ok {
			t.Fatalf(
				"ListServices() filtered returned unexpected service ID %d",
				service.ID,
			)
		}
		delete(wantPlatformIDs, service.ID)
	}
	if len(wantPlatformIDs) != 0 {
		t.Fatalf(
			"ListServices() filtered missed service IDs %#v",
			wantPlatformIDs,
		)
	}
}

func TestServiceRepositoryIntegrationConcurrentDuplicateCreate(t *testing.T) {
	_, teamRepository, pool, ctx := setupServiceRepositoryIntegration(t)

	team := createIntegrationTeam(t, ctx, teamRepository, "platform", "Platform")

	type result struct {
		service catalog.Service
		err     error
	}

	repositories := []*ServiceRepository{
		NewServiceRepository(pool),
		NewServiceRepository(pool),
	}

	start := make(chan struct{})
	results := make(chan result, len(repositories))

	var group sync.WaitGroup
	for _, repository := range repositories {
		group.Add(1)

		go func(repository *ServiceRepository) {
			defer group.Done()

			<-start

			detail, err := repository.CreateService(
				ctx,
				catalog.CreateServiceInput{
					TeamID: team.ID,
					Slug:   "catalog-api",
					Name:   "Catalog API",
				},
			)
			results <- result{
				service: detail.Service,
				err:     err,
			}
		}(repository)
	}

	close(start)
	group.Wait()
	close(results)

	successes := 0
	conflicts := 0

	for result := range results {
		switch {
		case result.err == nil:
			successes++

			if result.service.ID <= 0 {
				t.Errorf(
					"concurrent CreateService() ID = %d, want positive",
					result.service.ID,
				)
			}

		case errors.Is(result.err, catalog.ErrConflict):
			conflicts++

		default:
			t.Errorf(
				"concurrent CreateService() error = %v, want ErrConflict",
				result.err,
			)
		}
	}

	if successes != 1 {
		t.Fatalf("concurrent CreateService() successes = %d, want 1", successes)
	}
	if conflicts != 1 {
		t.Fatalf("concurrent CreateService() conflicts = %d, want 1", conflicts)
	}

	var count int
	if err := pool.QueryRow(
		ctx,
		"SELECT count(*) FROM services WHERE team_id = $1 AND slug = $2",
		team.ID,
		"catalog-api",
	).Scan(&count); err != nil {
		t.Fatalf("count concurrently created services: %v", err)
	}
	if count != 1 {
		t.Fatalf("concurrently created service count = %d, want 1", count)
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

func setupServiceRepositoryIntegration(
	t *testing.T,
) (*ServiceRepository, *TeamRepository, *pgxpool.Pool, context.Context) {
	t.Helper()

	pool := openIntegrationPool(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	resetCatalogTables(t, ctx, pool)

	return NewServiceRepository(pool), NewTeamRepository(pool), pool, ctx
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

func createIntegrationService(
	t *testing.T,
	ctx context.Context,
	repository *ServiceRepository,
	teamID int64,
	slug string,
	name string,
) catalog.Service {
	t.Helper()

	detail, err := repository.CreateService(
		ctx,
		catalog.CreateServiceInput{
			TeamID: teamID,
			Slug:   slug,
			Name:   name,
		},
	)
	if err != nil {
		t.Fatalf("CreateService(%q) error = %v", slug, err)
	}

	return detail.Service
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

func assertServicesInDescendingCursorOrder(
	t *testing.T,
	services []catalog.Service,
) {
	t.Helper()

	for index := 0; index+1 < len(services); index++ {
		current := services[index]
		next := services[index+1]

		if current.CreatedAt.Before(next.CreatedAt) {
			t.Fatalf(
				"Services are not ordered by created_at DESC: %s before %s",
				current.CreatedAt,
				next.CreatedAt,
			)
		}
		if current.CreatedAt.Equal(next.CreatedAt) && current.ID <= next.ID {
			t.Fatalf(
				"Services with equal created_at are not ordered by id DESC: %d before %d",
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
