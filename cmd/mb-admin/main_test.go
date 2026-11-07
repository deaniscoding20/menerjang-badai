package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func provisionArgs() []string {
	return []string{"provision", "--confirm", "--issuer", "https://idp.example/application/o/dashboard/", "--subject", "immutable-subject", "--display-name", "Local Owner", "--org-name", "Local Organization", "--brand-name", "Local Brand"}
}

func TestParseAdminCommands(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"--help"}, "help"},
		{[]string{"provision", "--help"}, "help"},
		{[]string{"db-ping"}, "db-ping"},
		{[]string{"migrate", "--confirm"}, "migrate"},
		{provisionArgs(), "provision"},
	} {
		got, err := parseCommand(tc.args)
		if err != nil || got.command != tc.want {
			t.Fatalf("command %q: got %q, %v", tc.want, got.command, err)
		}
	}
}

func TestParseAdminRejectsUnsafeOrIncompleteArguments(t *testing.T) {
	cases := [][]string{
		nil, {"serve"}, {"db-ping", "extra"}, {"migrate"}, {"migrate", "--confirm", "extra"},
		{"provision", "--confirm"}, {"--password=credential-canary"}, {"migrate", "--confirm=credential-canary"},
	}
	for _, change := range []struct{ key, value string }{
		{"--issuer", "http://idp.example"}, {"--issuer", "https://credential-canary@idp.example"},
		{"--issuer", "https://idp.example/?secret=credential-canary"},
		{"--subject", ""}, {"--subject", " subject "}, {"--display-name", "\n"},
		{"--brand-name", strings.Repeat("x", 121)},
	} {
		args := provisionArgs()
		for i := range args {
			if args[i] == change.key {
				args[i+1] = change.value
			}
		}
		cases = append(cases, args)
	}
	for _, args := range cases {
		if _, err := parseCommand(args); err == nil || strings.Contains(err.Error(), "credential-canary") {
			t.Fatal("invalid arguments must fail with a fixed safe error")
		}
	}
}

func TestAdminMissingConfigFailsBeforeDBAccess(t *testing.T) {
	for _, key := range []string{"MB_DATABASE_URL_FILE", "MB_DATABASE_URL", "PGPASSWORD", "MB_DATABASE_PASSWORD", "MB_OIDC_CLIENT_SECRET"} {
		t.Setenv(key, "")
	}
	var out, stderr bytes.Buffer
	if code := run(context.Background(), []string{"db-ping"}, &out, &stderr); code == 0 || out.Len() != 0 || !strings.Contains(stderr.String(), "database configuration invalid") {
		t.Fatal("database configuration should fail closed")
	}
}

func TestAdminRawSecretEnvIsRejectedSafely(t *testing.T) {
	for _, key := range []string{"MB_DATABASE_URL", "PGPASSWORD", "MB_DATABASE_PASSWORD", "MB_OIDC_CLIENT_SECRET"} {
		t.Run(key, func(t *testing.T) {
			t.Setenv(key, "credential-canary")
			var out, stderr bytes.Buffer
			if code := run(context.Background(), []string{"db-ping"}, &out, &stderr); code != 1 || out.Len() != 0 || strings.Contains(stderr.String(), "credential-canary") || !strings.Contains(stderr.String(), "database configuration invalid") {
				t.Fatal("raw secret environment value must be rejected without being echoed")
			}
		})
	}
}

func TestAdminHelpHasNoDependencyAccess(t *testing.T) {
	var out, stderr bytes.Buffer
	if code := run(context.Background(), []string{"--help"}, &out, &stderr); code != 0 || stderr.Len() != 0 || !strings.Contains(out.String(), "--confirm") {
		t.Fatal("expected help without database/configuration access")
	}
}
