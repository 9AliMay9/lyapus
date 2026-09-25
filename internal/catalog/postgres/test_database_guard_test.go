package postgres

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func validateIntegrationDatabaseName(name string) error {
	if name == "lyapus_dev_test" {
		return fmt.Errorf("legacy development database is not a test target")
	}

	if !strings.HasSuffix(name, "_test") {
		return fmt.Errorf("integration database name must end with _test")
	}

	return nil
}

func TestValidateIntegrationDatabaseName(t *testing.T) {
	tests := []struct {
		label   string
		name    string
		wantErr bool
	}{
		{
			label:   "new_development_database",
			name:    "lyapus_dev",
			wantErr: true,
		},
		{
			label:   "legacy_development_database",
			name:    "lyapus_dev_test",
			wantErr: true,
		},
		{
			label:   "empty",
			name:    "",
			wantErr: true,
		},
		{
			label:   "missing_suffix",
			name:    "lyapus",
			wantErr: true,
		},
		{
			label: "integration_database",
			name:  "lyapus_integration_test",
		},
		{
			label: "ci_database",
			name:  "lyapus_verify_test",
		},
		{
			label: "constraint_database",
			name:  "lyapus_constraints_test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			err := validateIntegrationDatabaseName(tt.name)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"validateIntegrationDatabaseName(%q) error = %v, wantErr = %v",
					tt.name,
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestValidateParsedIntegrationDatabaseName(t *testing.T) {
	tests := []struct {
		label        string
		databaseURL  string
		wantDatabase string
		wantErr      bool
	}{
		{
			label:        "ordinary_test_database",
			databaseURL:  "postgres://lyapus:lyapus@127.0.0.1:1/lyapus_integration_test?sslmode=disable",
			wantDatabase: "lyapus_integration_test",
		},
		{
			label:        "dbname_overrides_to_legacy_development",
			databaseURL:  "postgres://lyapus:lyapus@127.0.0.1:1/lyapus_integration_test?sslmode=disable&dbname=lyapus_dev_test",
			wantDatabase: "lyapus_dev_test",
			wantErr:      true,
		},
		{
			label:        "database_overrides_to_new_development",
			databaseURL:  "postgres://lyapus:lyapus@127.0.0.1:1/lyapus_integration_test?sslmode=disable&database=lyapus_dev",
			wantDatabase: "lyapus_dev",
			wantErr:      true,
		},
		{
			label:        "encoded_legacy_development_name",
			databaseURL:  "postgres://lyapus:lyapus@127.0.0.1:1/lyapus%5Fdev%5Ftest?sslmode=disable",
			wantDatabase: "lyapus_dev_test",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			poolConfig, err := pgxpool.ParseConfig(tt.databaseURL)
			if err != nil {
				t.Fatal("parse fixture configuration failed")
			}

			databaseName := poolConfig.ConnConfig.Database
			if databaseName != tt.wantDatabase {
				t.Fatalf(
					"parsed database = %q, want %q",
					databaseName,
					tt.wantDatabase,
				)
			}

			err = validateIntegrationDatabaseName(databaseName)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"validation error = %v, wantErr = %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}
