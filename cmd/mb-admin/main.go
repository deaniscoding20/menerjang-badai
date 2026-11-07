package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/menerjangbadai/dashboard/db"
	"github.com/menerjangbadai/dashboard/internal/domain"
	"github.com/menerjangbadai/dashboard/internal/platform"
)

var buildSHA string

const usage = `Usage:
  mb-admin db-ping
  mb-admin migrate --confirm
  mb-admin provision --confirm --issuer HTTPS_ISSUER --subject EXACT_SUBJECT \
    --display-name NAME --org-name NAME --brand-name NAME

Database commands read MB_DATABASE_URL_FILE, never a command-line DSN.
Migrate is an explicit forward-only operation. Provision creates a NEW organization,
NEW owner user and NEW brand with membership/access in a single transaction.
Provision requires full runtime configuration and the exact configured OIDC issuer.
Existing identities are rejected, never promoted or reassigned. No password/login bypass.
See docs/14-local-development.md. Use --help for this message.
`

var errArguments = errors.New("invalid or incomplete command arguments; use --help")

type options struct {
	command                      string
	issuer, subject, displayName string
	orgName, brandName           string
}

func parseCommand(args []string) (options, error) {
	if len(args) == 0 {
		return options{}, errArguments
	}
	if len(args) == 1 && (args[0] == "help" || args[0] == "--help" || args[0] == "-h") {
		return options{command: "help"}, nil
	}
	o := options{command: args[0]}
	if o.command != "db-ping" && o.command != "migrate" && o.command != "provision" {
		return options{}, errArguments
	}
	if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
		return options{command: "help"}, nil
	}
	fs := flag.NewFlagSet("mb-admin", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // Flag diagnostics may echo untrusted argument values.
	confirm := false
	if o.command != "db-ping" {
		fs.BoolVar(&confirm, "confirm", false, "confirm a database write")
	}
	if o.command == "provision" {
		fs.StringVar(&o.issuer, "issuer", "", "exact OIDC issuer")
		fs.StringVar(&o.subject, "subject", "", "immutable OIDC subject")
		fs.StringVar(&o.displayName, "display-name", "", "new user display name")
		fs.StringVar(&o.orgName, "org-name", "", "new organization name")
		fs.StringVar(&o.brandName, "brand-name", "", "new brand name")
	}
	if err := fs.Parse(args[1:]); err != nil || fs.NArg() != 0 || (o.command != "db-ping" && !confirm) {
		return options{}, errArguments
	}
	if o.command == "provision" {
		u, err := url.Parse(o.issuer)
		if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || strings.ContainsAny(o.issuer, " \t\r\n") {
			return options{}, errArguments
		}
		if !validText(o.subject, 2048) || o.subject != strings.TrimSpace(o.subject) || !validText(o.displayName, 200) || !validText(o.orgName, 200) || !validText(o.brandName, 120) {
			return options{}, errArguments
		}
		o.displayName, o.orgName, o.brandName = strings.TrimSpace(o.displayName), strings.TrimSpace(o.orgName), strings.TrimSpace(o.brandName)
		if (domain.Brand{Name: o.brandName, Status: "active"}).Validate() != nil {
			return options{}, errArguments
		}
	}
	return o, nil
}

func validText(s string, max int) bool {
	return utf8.ValidString(s) && strings.TrimSpace(s) != "" && utf8.RuneCountInString(s) <= max && strings.IndexFunc(s, unicode.IsControl) == -1
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	os.Exit(run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, out, stderr io.Writer) int {
	o, err := parseCommand(args)
	if err != nil {
		fmt.Fprintln(stderr, "mb-admin:", err)
		return 2
	}
	if o.command == "help" {
		fmt.Fprint(out, usage)
		return 0
	}
	dsn, err := platform.DatabaseURL()
	if err != nil {
		fmt.Fprintln(stderr, "mb-admin: database configuration invalid; use a private MB_DATABASE_URL_FILE")
		return 1
	}
	if o.command == "provision" {
		c, err := platform.LoadRuntimeConfig(buildSHA)
		if err != nil {
			fmt.Fprintln(stderr, "mb-admin:", err)
			return 1
		}
		if c.Issuer != o.issuer {
			fmt.Fprintln(stderr, "mb-admin: issuer must exactly match configured OIDC issuer")
			return 1
		}
		dsn = c.DatabaseURL
	}
	operation, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	pool, err := pgxpool.New(operation, dsn)
	if err != nil {
		fmt.Fprintln(stderr, "mb-admin: database configuration rejected")
		return 1
	}
	defer pool.Close()
	ping, done := context.WithTimeout(operation, 5*time.Second)
	err = pool.Ping(ping)
	done()
	if err != nil {
		fmt.Fprintln(stderr, "mb-admin: database unavailable")
		return 1
	}
	switch o.command {
	case "db-ping":
		fmt.Fprintln(out, "database reachable (schema not checked)")
	case "migrate":
		if err := db.Migrate(operation, pool); err != nil {
			fmt.Fprintln(stderr, "mb-admin: migration failed; inspect database privately")
			return 1
		}
		fmt.Fprintln(out, "migrations applied")
	case "provision":
		result, err := provision(operation, pool, o)
		if err != nil {
			fmt.Fprintln(stderr, "mb-admin: provisioning failed; existing identities are never promoted; inspect database privately")
			return 1
		}
		if json.NewEncoder(out).Encode(result) != nil {
			fmt.Fprintln(stderr, "mb-admin: provisioning committed but result output failed; inspect database before retrying")
			return 1
		}
	}
	return 0
}

type provisionResult struct {
	OrgID   string `json:"org_id"`
	UserID  string `json:"user_id"`
	BrandID string `json:"brand_id"`
}

// Provision only creates fresh records; external_identities' primary key rejects
// existing issuer/subject mappings. All writes roll back together on conflict.
func provision(ctx context.Context, pool *pgxpool.Pool, o options) (provisionResult, error) {
	result := provisionResult{OrgID: uuid.NewString(), UserID: uuid.NewString(), BrandID: uuid.NewString()}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return provisionResult{}, err
	}
	defer tx.Rollback(ctx)
	commands := []struct {
		sql  string
		args []any
	}{
		{"INSERT INTO organizations(id,name) VALUES($1,$2)", []any{result.OrgID, o.orgName}},
		{"INSERT INTO users(id,display_name) VALUES($1,$2)", []any{result.UserID, o.displayName}},
		{"INSERT INTO external_identities(issuer,subject,user_id) VALUES($1,$2,$3)", []any{o.issuer, o.subject, result.UserID}},
		{"INSERT INTO organization_members(org_id,user_id) VALUES($1,$2)", []any{result.OrgID, result.UserID}},
		{"INSERT INTO user_roles(org_id,user_id,role_id) VALUES($1,$2,'owner')", []any{result.OrgID, result.UserID}},
		{"INSERT INTO brands(org_id,id,name) VALUES($1,$2,$3)", []any{result.OrgID, result.BrandID, o.brandName}},
		{"INSERT INTO user_brand_access(org_id,brand_id,user_id) VALUES($1,$2,$3)", []any{result.OrgID, result.BrandID, result.UserID}},
		{"INSERT INTO audit_logs(org_id,user_id,event) VALUES($1,$2,'identity.provisioned')", []any{result.OrgID, result.UserID}},
		{"INSERT INTO audit_logs(org_id,user_id,event,resource_id) VALUES($1,$2,'brand.created',$3)", []any{result.OrgID, result.UserID, result.BrandID}},
	}
	for _, c := range commands {
		if _, err := tx.Exec(ctx, c.sql, c.args...); err != nil {
			return provisionResult{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return provisionResult{}, err
	}
	return result, nil
}
