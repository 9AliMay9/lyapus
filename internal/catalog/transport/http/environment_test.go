package cataloghttp

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/9AliMay9/lyapus/internal/catalog"
)

func TestCreateEnvironmentInputFromRequest(t *testing.T) {
	got := createEnvironmentInputFromRequest(createEnvironmentRequest{
		ServiceID: 7,
		Slug:      "production",
		Name:      "Production",
	})

	if got.ServiceID != 7 {
		t.Fatalf("ServiceID = %d, want 7", got.ServiceID)
	}
	if got.Slug != "production" {
		t.Fatalf("Slug = %q, want production", got.Slug)
	}
	if got.Name != "Production" {
		t.Fatalf("Name = %q, want Production", got.Name)
	}
}

func TestUpdateEnvironmentRequestToCatalogInput(t *testing.T) {
	t.Run("provided fields", func(t *testing.T) {
		var request updateEnvironmentRequest
		if err := json.Unmarshal(
			[]byte(`{"slug":"staging","name":"Staging"}`),
			&request,
		); err != nil {
			t.Fatalf("unmarshal update environment request: %v", err)
		}

		got := updateEnvironmentInputFromRequest(request)

		if got.Slug == nil || *got.Slug != "staging" {
			t.Fatalf("Slug = %#v, want staging", got.Slug)
		}
		if got.Name == nil || *got.Name != "Staging" {
			t.Fatalf("Name = %#v, want Staging", got.Name)
		}
	})

	t.Run("omitted fields", func(t *testing.T) {
		var request updateEnvironmentRequest
		if err := json.Unmarshal([]byte(`{}`), &request); err != nil {
			t.Fatalf("unmarshal update environment request: %v", err)
		}

		got := updateEnvironmentInputFromRequest(request)

		if got.Slug != nil || got.Name != nil {
			t.Fatalf("input = %#v, want no values", got)
		}
	})

	t.Run("rejects null string field", func(t *testing.T) {
		var request updateEnvironmentRequest
		err := json.Unmarshal([]byte(`{"name":null}`), &request)
		if err == nil {
			t.Fatal("unmarshal update environment request error = nil, want error")
		}
	})
}

func TestEnvironmentResponseFromCatalog(t *testing.T) {
	createdAt := time.Date(
		2026,
		time.September,
		5,
		8,
		30,
		0,
		0,
		time.FixedZone("CST", 8*60*60),
	)
	updatedAt := createdAt.Add(time.Minute)

	got := environmentResponseFromCatalog(catalog.Environment{
		ID:        11,
		ServiceID: 7,
		Slug:      "production",
		Name:      "Production",
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	})

	if got.ID != 11 || got.ServiceID != 7 {
		t.Fatalf("response = %#v, want ID 11 and ServiceID 7", got)
	}
	if got.Slug != "production" || got.Name != "Production" {
		t.Fatalf("response = %#v, want production / Production", got)
	}
	if !got.CreatedAt.Equal(createdAt.UTC()) {
		t.Fatalf("CreatedAt = %s, want %s", got.CreatedAt, createdAt.UTC())
	}
	if !got.UpdatedAt.Equal(updatedAt.UTC()) {
		t.Fatalf("UpdatedAt = %s, want %s", got.UpdatedAt, updatedAt.UTC())
	}
}
