package cataloghttp

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/9AliMay9/lyapus/internal/catalog"
)

func TestCreateServiceInputFromRequest(t *testing.T) {
	description := "Service catalog API"

	got := createServiceInputFromRequest(createServiceRequest{
		TeamID:      7,
		Slug:        "catalog-api",
		Name:        "Catalog API",
		Description: &description,
		Environments: []createEnvironmentRequest{
			{
				Slug: "staging",
				Name: "Staging",
			},
			{
				Slug: "production",
				Name: "Production",
			},
		},
	})

	if got.TeamID != 7 {
		t.Fatalf("TeamID = %d, want 7", got.TeamID)
	}
	if got.Slug != "catalog-api" || got.Name != "Catalog API" {
		t.Fatalf("input = %#v, want catalog-api / Catalog API", got)
	}
	if got.Description == nil || *got.Description != description {
		t.Fatalf("Description = %#v, want %q", got.Description, description)
	}
	if len(got.Environments) != 2 {
		t.Fatalf("environment count = %d, want 2", len(got.Environments))
	}
	if got.Environments[0].Slug != "staging" {
		t.Fatalf(
			"first environment slug = %q, want staging",
			got.Environments[0].Slug,
		)
	}
	if got.Environments[1].Name != "Production" {
		t.Fatalf(
			"second environment name = %q, want Production",
			got.Environments[1].Name,
		)
	}
}

func TestServiceResponsesFromCatalog(t *testing.T) {
	createdAt := time.Date(
		2026,
		time.August,
		20,
		12,
		0,
		0,
		0,
		time.FixedZone("CST", 8*60*60),
	)
	updatedAt := createdAt.Add(time.Hour)
	description := "Service catalog API"

	detail := catalog.ServiceDetail{
		Service: catalog.Service{
			ID:          9,
			TeamID:      7,
			Slug:        "catalog-api",
			Name:        "Catalog API",
			Description: &description,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		},
		Environments: []catalog.Environment{
			{
				ID:        11,
				ServiceID: 9,
				Slug:      "production",
				Name:      "Production",
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			},
		},
	}

	listItem := serviceResponseFromCatalog(detail.Service)
	if listItem.ID != 9 || listItem.TeamID != 7 {
		t.Fatalf("service response IDs = (%d, %d), want (9, 7)", listItem.ID, listItem.TeamID)
	}
	if listItem.Description == nil || *listItem.Description != description {
		t.Fatalf(
			"service response Description = %#v, want %q",
			listItem.Description,
			description,
		)
	}
	if !listItem.CreatedAt.Equal(createdAt.UTC()) {
		t.Fatalf(
			"service response CreatedAt = %s, want %s",
			listItem.CreatedAt,
			createdAt.UTC(),
		)
	}
	if listItem.CreatedAt.Location() != time.UTC {
		t.Fatalf(
			"service response CreatedAt location = %s, want UTC",
			listItem.CreatedAt.Location(),
		)
	}

	got := serviceDetailResponseFromCatalog(detail)
	if got.ID != 9 || got.TeamID != 7 {
		t.Fatalf("detail response IDs = (%d, %d), want (9, 7)", got.ID, got.TeamID)
	}
	if got.Slug != "catalog-api" || got.Name != "Catalog API" {
		t.Fatalf("detail response = %#v, want catalog-api / Catalog API", got)
	}
	if len(got.Environments) != 1 {
		t.Fatalf("detail environment count = %d, want 1", len(got.Environments))
	}
	if got.Environments[0].ID != 11 || got.Environments[0].ServiceID != 9 {
		t.Fatalf(
			"environment response IDs = (%d, %d), want (11, 9)",
			got.Environments[0].ID,
			got.Environments[0].ServiceID,
		)
	}
	if got.Environments[0].CreatedAt.Location() != time.UTC {
		t.Fatalf(
			"environment response CreatedAt location = %s, want UTC",
			got.Environments[0].CreatedAt.Location(),
		)
	}

	emptyDetail := serviceDetailResponseFromCatalog(catalog.ServiceDetail{})
	encoded, err := json.Marshal(emptyDetail)
	if err != nil {
		t.Fatalf("marshal empty detail response: %v", err)
	}
	if !strings.Contains(string(encoded), `"environments":[]`) {
		t.Fatalf(
			"empty detail response JSON = %s, want environments as []",
			encoded,
		)
	}
}
