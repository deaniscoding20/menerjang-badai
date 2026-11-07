package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunRejectsMissingConfigurationBeforeStartup(t *testing.T) {
	for _, key := range []string{"MB_COMMIT_SHA", "MB_PUBLIC_URL", "MB_OIDC_ISSUER", "MB_OIDC_CLIENT_ID", "MB_DATABASE_URL_FILE", "MB_OIDC_CLIENT_SECRET_FILE"} {
		t.Setenv(key, "")
	}
	for _, args := range [][]string{nil, {"serve"}, {"check-config"}} {
		var out, stderr bytes.Buffer
		if code := run(context.Background(), args, &out, &stderr); code == 0 {
			t.Fatal("missing configuration succeeded")
		}
		if !strings.Contains(stderr.String(), "configuration invalid") || out.Len() != 0 {
			t.Fatal("missing safe configuration error")
		}
	}
}

func TestRunHelpDoesNotRequireConfig(t *testing.T) {
	var out, stderr bytes.Buffer
	if code := run(context.Background(), []string{"--help"}, &out, &stderr); code != 0 || !strings.Contains(out.String(), "check-config") || stderr.Len() != 0 {
		t.Fatal("help should succeed without configuration")
	}
}

func TestHTTPDiagnosticsNeverExposeRawInput(t *testing.T) {
	var out bytes.Buffer
	raw := []byte("postgres://credential-canary@host/db")
	n, err := (safeHTTPLog{out: &out}).Write(raw)
	if err != nil || n != len(raw) || strings.Contains(out.String(), "credential-canary") || !strings.Contains(out.String(), "diagnostic suppressed") {
		t.Fatal("HTTP diagnostics must be replaced with a fixed safe message")
	}
}

func TestParseCommand(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"default serve", nil, "serve"},
		{"explicit serve", []string{"serve"}, "serve"},
		{"config check", []string{"check-config"}, "check-config"},
		{"help", []string{"--help"}, "help"},
		{"command help", []string{"serve", "--help"}, "help"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseCommand(tc.args)
			if err != nil || got != tc.want {
				t.Fatalf("got %q, %v; want %q", got, err, tc.want)
			}
		})
	}
	for _, args := range [][]string{{"migrate"}, {"unknown"}, {"serve", "extra"}, {"--dsn=credential-canary"}} {
		if _, err := parseCommand(args); err == nil || strings.Contains(err.Error(), "credential-canary") {
			t.Fatal("invalid input must return a fixed safe error")
		}
	}
}
