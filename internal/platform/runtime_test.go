package platform

import (
	"strings"
	"testing"
)

func clearRuntimeEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"MB_COMMIT_SHA", "MB_HTTP_ADDR", "MB_PUBLIC_URL", "MB_OIDC_ISSUER", "MB_OIDC_CLIENT_ID", "MB_WEB_DIR",
		"MB_DATABASE_URL_FILE", "MB_OIDC_CLIENT_SECRET_FILE", "MB_DATABASE_URL", "PGPASSWORD", "MB_DATABASE_PASSWORD", "MB_OIDC_CLIENT_SECRET",
	} {
		t.Setenv(key, "")
	}
}

func TestLoadRuntimeConfigFailsClosed(t *testing.T) {
	clearRuntimeEnv(t)
	for _, sha := range []string{"", strings.Repeat("a", 40)} {
		c, err := LoadRuntimeConfig(sha)
		if err == nil || c != (Config{}) {
			t.Fatal("missing configuration must return an error and no partial config")
		}
		if err.Error() != "configuration invalid; see docs/14-local-development.md" {
			t.Fatalf("unexpected safe error: %s", err)
		}
	}
}

func TestRuntimeAddressAndDatabaseSyntax(t *testing.T) {
	for _, tc := range []struct {
		name  string
		addr  string
		dsn   string
		valid bool
	}{
		{"default loopback", "127.0.0.1:12023", "postgres://runtime-test@127.0.0.1:1/runtime-test?sslmode=disable", true},
		{"ipv6 loopback", "[::1]:12023", "postgres://runtime-test@127.0.0.1:1/runtime-test", true},
		{"reserved website port", "127.0.0.1:2023", "postgres://runtime-test@127.0.0.1:1/runtime-test", false},
		{"non-loopback", "0.0.0.0:12023", "postgres://runtime-test@127.0.0.1:1/runtime-test", false},
		{"ephemeral port", "127.0.0.1:0", "postgres://runtime-test@127.0.0.1:1/runtime-test", false},
		{"invalid port", "127.0.0.1:65536", "postgres://runtime-test@127.0.0.1:1/runtime-test", false},
		{"invalid DSN", "127.0.0.1:12023", "postgres://credential-canary@127.0.0.1:invalid/test", false},
		{"empty DSN", "127.0.0.1:12023", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateRuntimeConfig(Config{Addr: tc.addr, DatabaseURL: tc.dsn})
			if (err == nil) != tc.valid {
				t.Fatalf("validation result: %v", err)
			}
			if err != nil && strings.Contains(err.Error(), "credential-canary") {
				t.Fatal("configuration error exposed DSN contents")
			}
		})
	}
}

func TestLoadRuntimeConfigDoesNotEchoInvalidValues(t *testing.T) {
	clearRuntimeEnv(t)
	t.Setenv("MB_COMMIT_SHA", strings.Repeat("a", 40))
	t.Setenv("MB_HTTP_ADDR", "credential-canary")
	_, err := LoadRuntimeConfig("")
	if err == nil || strings.Contains(err.Error(), "credential-canary") {
		t.Fatal("invalid configuration must not echo supplied values")
	}
}
