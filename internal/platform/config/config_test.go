package config

import "testing"

func TestLoadUsesDefaultHTTPAddr(t *testing.T) {
	t.Setenv("LYAPUS_HTTP_ADDR", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTPAddr != defaultHTTPAddr {
		t.Fatalf("HTTPAddr = %q, want %q", cfg.HTTPAddr, defaultHTTPAddr)
	}
}

func TestLoadUsesConfiguredHTTPAddr(t *testing.T) {
	t.Setenv("LYAPUS_HTTP_ADDR", "127.0.0.1:9090")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HTTPAddr != "127.0.0.1:9090" {
		t.Fatalf("HTTPAddr = %q, want %q", cfg.HTTPAddr, "127.0.0.1:9090")
	}
}

func TestLoadRejectsInvalidHTTPAddr(t *testing.T) {
	t.Setenv("LYAPUS_HTTP_ADDR", "127.0.0.1")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want an error")
	}
}
