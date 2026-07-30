package config

import "testing"

const testDatabaseURL = "postgres://lyapus:lyapus@127.0.0.1:5432/lyapus?sslmode=disable"

func TestLoadUsesDefaultHTTPAddr(t *testing.T) {
	t.Setenv("LYAPUS_HTTP_ADDR", "")
	t.Setenv("LYAPUS_DATABASE_URL", testDatabaseURL)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTPAddr != defaultHTTPAddr {
		t.Fatalf("HTTPAddr = %q, want %q", cfg.HTTPAddr, defaultHTTPAddr)
	}
	if cfg.DatabaseURL != testDatabaseURL {
		t.Fatalf("DatabaseURL = %q, want %q", cfg.DatabaseURL, testDatabaseURL)
	}
}

func TestLoadUsesConfiguredHTTPAddr(t *testing.T) {
	t.Setenv("LYAPUS_HTTP_ADDR", "127.0.0.1:9090")
	t.Setenv("LYAPUS_DATABASE_URL", testDatabaseURL)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HTTPAddr != "127.0.0.1:9090" {
		t.Fatalf("HTTPAddr = %q, want %q", cfg.HTTPAddr, "127.0.0.1:9090")
	}
	if cfg.DatabaseURL != testDatabaseURL {
		t.Fatalf("DatabaseURL = %q, want %q", cfg.DatabaseURL, testDatabaseURL)
	}
}

func TestLoadRejectsInvalidHTTPAddr(t *testing.T) {
	t.Setenv("LYAPUS_HTTP_ADDR", "127.0.0.1")
	t.Setenv("LYAPUS_DATABASE_URL", testDatabaseURL)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want an error")
	}
}

func TestLoadRejectsMissingDatabaseURL(t *testing.T) {
	t.Setenv("LYAPUS_DATABASE_URL", "")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want an error")
	}
}

func TestLoadRejectsUnsupportedDatabaseURLScheme(t *testing.T) {
	t.Setenv("LYAPUS_DATABASE_URL", "mysql://lyapus:lyapus@127.0.0.1:3306/lyapus")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want an error")
	}
}

func TestLoadAcceptsPostgreDatabaseURL(t *testing.T) {
	databaseURL := "postgresql://lyapus:lyapus@127.0.0.1:5432/lyapus?sslmode=disable"
	t.Setenv("LYAPUS_DATABASE_URL", databaseURL)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.DatabaseURL != databaseURL {
		t.Fatalf("DatabaseURL = %q, want %q", cfg.DatabaseURL, databaseURL)
	}
}
