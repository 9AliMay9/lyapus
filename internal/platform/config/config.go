package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const defaultHTTPAddr = "127.0.0.1:8080"

type Config struct {
	HTTPAddr    string
	DatabaseURL string
}

func Load() (Config, error) {
	httpAddr := os.Getenv("LYAPUS_HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = defaultHTTPAddr
	}

	if err := validateHTTPAddr(httpAddr); err != nil {
		return Config{}, fmt.Errorf("validate LYAPUS_HTTP_ADDR: %w", err)
	}

	databaseURL := os.Getenv("LYAPUS_DATABASE_URL")
	if err := validateDatabaseURL(databaseURL); err != nil {
		return Config{}, fmt.Errorf("validate LYAPUS_DATABASE_URL: %w", err)
	}

	return Config{
		HTTPAddr:    httpAddr,
		DatabaseURL: databaseURL,
	}, nil
}

func validateHTTPAddr(addr string) error {
	if strings.TrimSpace(addr) == "" {
		return fmt.Errorf("must not be empty")
	}

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("must be in host:port form: %w", err)
	}
	if host == "" {
		return fmt.Errorf("host must not be empty")
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return fmt.Errorf("port must be an integer from 1 through 65535")
	}

	return nil
}

func validateDatabaseURL(rawURL string) error {
	if strings.TrimSpace(rawURL) == "" {
		return fmt.Errorf("must not be empty")
	}

	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("must be a valid PostgreSQL connection URL: %w", err)
	}

	switch parsedURL.Scheme {
	case "postgres", "postgresql":
		return nil
	default:
		return fmt.Errorf("scheme must be postgres or postgresql")
	}
}
