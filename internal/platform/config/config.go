package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const defaultHTTPAddr = "127.0.0.1:8080"

type Config struct {
	HTTPAddr string
}

func Load() (Config, error) {
	httpAddr := os.Getenv("LYAPUS_HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = defaultHTTPAddr
	}

	if err := validateHTTPAddr(httpAddr); err != nil {
		return Config{}, fmt.Errorf("validate LYAPUS_HTTP_ADDR: %w", err)
	}

	return Config{HTTPAddr: httpAddr}, nil
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
