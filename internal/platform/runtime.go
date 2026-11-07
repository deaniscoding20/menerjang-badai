package platform

import (
	"errors"
	"net"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

var errRuntimeConfig = errors.New("configuration invalid; see docs/14-local-development.md")

// LoadRuntimeConfig preserves strict production authentication requirements in development.
// Returning no partial configuration keeps accidental logging from exposing loaded secrets.
func LoadRuntimeConfig(buildSHA string) (Config, error) {
	c, err := Load(buildSHA)
	if err != nil {
		return Config{}, errRuntimeConfig
	}
	if err := validateRuntimeConfig(c); err != nil {
		return Config{}, errRuntimeConfig
	}
	return c, nil
}

func validateRuntimeConfig(c Config) error {
	host, rawPort, err := net.SplitHostPort(c.Addr)
	port, portErr := strconv.Atoi(rawPort)
	ip := net.ParseIP(host)
	if err != nil || ip == nil || !ip.IsLoopback() || portErr != nil || port < 1 || port > 65535 || port == 2023 {
		return errRuntimeConfig
	}
	if strings.TrimSpace(c.DatabaseURL) == "" {
		return errRuntimeConfig
	}
	// Parsing only: no connection, discovery, query or migration is performed.
	if _, err := pgxpool.ParseConfig(c.DatabaseURL); err != nil {
		return errRuntimeConfig
	}
	return nil
}
