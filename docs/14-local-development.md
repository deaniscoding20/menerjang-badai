# Local application development

## Boundaries and evidence

The Go entrypoints are `cmd/mb-dashboard` and `cmd/mb-admin`. This is a development
startup path, **not a production deployment or an authentication bypass**. It does
not change a proxy, firewall, service, database, Hermes profile/bot or website2023.
The application binds `127.0.0.1:12023` by default; runtime validation rejects
non-loopback addresses, port 0 and the reserved website port 2023.

Short CLI parsing/configuration unit tests and Go builds are the local evidence.
Database migrations/provisioning, a running listener, shutdown under requests,
Authentik discovery/login and browser flows still require an approved isolated
environment. They have not been verified by these unit tests. Do not use a
production DSN as a fallback. Docker is not installed in the current development
environment; an isolated Postgres and a live approved Authentik client have not
been supplied. Do not install or start services merely to make a test pass.

## Prerequisites

- Go 1.27.3 (current host: `/home/ubuntu/.local/lib/go/bin/go`). Dependencies must
  already be cached for the offline commands below; otherwise arrange an approved
  dependency download separately. No changes to `go.mod` are needed by these CLIs.
- For live execution only: a disposable development PostgreSQL database with an
  appropriately scoped role, never a production or vendor service database.
- A real OIDC client with the exact HTTPS issuer, client ID and asymmetric RS256
  signing/JWKS. Register exactly `MB_PUBLIC_URL/auth/callback` as the redirect URI.
  The application uses authorization code + PKCE and verifies issuer, audience,
  authorized party and nonce. Email/display name is not an identity mapping.
- An already approved HTTPS endpoint/proxy to the loopback application for browser
  login. Secure `__Host-` cookies and HTTPS public/issuer URLs remain mandatory in
  development. Bare HTTP on loopback is not a substitute login flow. This guide
  does not authorize creating/changing proxy or public DNS configuration.
- A real frontend build under `web/dist` (or `MB_WEB_DIR`) for the UI. Backend
  startup does not compile the frontend; missing static files return 404.

## Configuration

Supply nonsecret settings in the process environment. Supply secrets **only by
reference to existing private files outside the repository**, provisioned through
an approved secret workflow. Do not paste secrets into a terminal command, chat,
CLI flags, source files or this document. These commands do not create any secret.

| Variable | Requirement |
| --- | --- |
| `MB_COMMIT_SHA` | Exact 40-character lowercase commit SHA; alternatively inject `main.buildSHA` at build time. No `dev`, `unknown` or short SHA fallback. |
| `MB_HTTP_ADDR` | Optional; default `127.0.0.1:12023`. Literal loopback IP and valid nonzero port required; 2023 forbidden. |
| `MB_PUBLIC_URL` | Explicit HTTPS origin with no path, credentials, query or fragment. |
| `MB_OIDC_ISSUER` | Explicit HTTPS issuer, including the provider's exact path/trailing slash. |
| `MB_OIDC_CLIENT_ID` | Required real provider client ID. |
| `MB_DATABASE_URL_FILE` | Path to an existing private regular file containing the PostgreSQL connection string. |
| `MB_OIDC_CLIENT_SECRET_FILE` | Path to an existing private regular file containing the real client secret. |
| `MB_WEB_DIR` | Optional static build directory; default `web/dist`, relative to the process working directory. |

Secret files must be regular, nonempty, at most 16 KiB, and have mode 0600 or
stricter (no group/other permissions). Raw `MB_DATABASE_URL`, `PGPASSWORD`,
`MB_DATABASE_PASSWORD` and `MB_OIDC_CLIENT_SECRET` environment variables are
rejected. Runtime errors are fixed safe messages; neither config objects nor
underlying driver/provider errors are printed. Inspect sensitive diagnostics only
through a separately approved private workflow. Do not enable shell tracing.

`mb-dashboard check-config` loads and validates the complete configuration and
parses the DSN, but does **not** connect, discover OIDC, migrate or listen. It does
not print the loaded configuration. An empty environment, absent files, invalid
SHA/URLs/address or invalid DSN returns nonzero. There is no implicit development
profile, mock login or local owner fallback.

## Build and short checks (no database or provider needed)

From the repository root:

```sh
GO=/home/ubuntu/.local/lib/go/bin/go
GOTOOLCHAIN=local GOPROXY=off "$GO" test ./cmd/mb-dashboard ./cmd/mb-admin ./internal/platform -count=1
GOTOOLCHAIN=local GOPROXY=off "$GO" build -o artifacts/mb-dashboard ./cmd/mb-dashboard
GOTOOLCHAIN=local GOPROXY=off "$GO" build -o artifacts/mb-admin ./cmd/mb-admin
./artifacts/mb-dashboard --help
./artifacts/mb-admin --help
```

`artifacts/` is ignored. An executable built without `main.buildSHA` requires
`MB_COMMIT_SHA` at runtime. To embed the verified release SHA, use
`-ldflags "-X main.buildSHA=$MB_COMMIT_SHA"` with **both** build commands after
setting the exact approved SHA. The environment override still wins. A SHA label
is not proof that an uncommitted working tree matches that commit.

## Explicit live steps — deferred until prerequisites are approved

These are operator commands to run later against the isolated development
database; documenting them is **not evidence they have run**.

1. Supply the settings above and existing private FILE references. Validate
   without network access:

   ```sh
   ./artifacts/mb-dashboard check-config
   ```

2. Check database reachability (no schema writes), then explicitly apply embedded
   forward migrations using the intended development migration role:

   ```sh
   ./artifacts/mb-admin db-ping
   ./artifacts/mb-admin migrate --confirm
   ```

   `db-ping` and `migrate` require only the FILE-based database configuration, not
   OIDC or a release SHA. All commands reject raw secret environment variables.
   Migration timeout is 60 seconds; migration checksums and the advisory lock are
   handled by `db.Migrate`. There is no automatic/down migration command. Prepare
   backup/restore and schema privileges before approving a real migration.

3. Obtain the **exact issuer and immutable subject** from the authorized provider
   administrator. Do not guess the subject from email or username. With full
   runtime configuration present, explicitly provision fresh local records:

   ```sh
   ./artifacts/mb-admin provision --confirm \
     --issuer "$MB_OIDC_ISSUER" --subject "$APPROVED_OIDC_SUBJECT" \
     --display-name 'Development Owner' --org-name 'Development Organization' \
     --brand-name 'Development Brand'
   ```

   The subject is an identifier, not a token or credential. Avoid this command in
   a shared shell where even identity metadata in argv/history is inappropriate.
   The issuer must exactly match configured `MB_OIDC_ISSUER`. This command does
   not create or verify an IdP account, issue a token or grant a bypass login.
   It atomically creates a new organization, new owner user, exact external
   identity mapping, organization membership/owner role, new active brand and
   active user-brand access; `identity.provisioned` and `brand.created` audit
   records are inserted. These are existing foundation-schema semantics.
   Success prints JSON with only `org_id`, `user_id`, `brand_id`.

   An existing `(issuer, subject)` is rejected by the identity primary key and
   the whole transaction rolls back. This is deliberately **not** an upsert,
   promotion, reassignment, existing-brand grant or invitation command. It also
   avoids the store's older `Bootstrap` helper because that helper hardcodes an
   organization name and does not provision a brand/access atomically. The login
   store currently requires exactly one active organization membership per
   external identity; multi-organization selection is not implemented. On an
   interrupted/uncertain commit, inspect the database privately before retrying.

4. Start explicitly (no command also means `serve`):

   ```sh
   ./artifacts/mb-dashboard serve
   ```

   Startup validates full config, creates a Postgres pool and pings it, discovers
   the real OIDC provider, then opens the configured loopback listener. **Serve
   never migrates, provisions or seeds.** Ping does not prove schema readiness;
   migrations must already be complete. Startup has a 20-second overall budget.
   A missing/unreachable dependency fails closed before listening. SIGINT/SIGTERM
   stop accepting requests and allow up to 10 seconds for HTTP graceful shutdown,
   then force-close if needed. HTTP read/header/write/idle timeouts are bounded.

Exit codes: `0` success/help, `2` invalid/incomplete CLI arguments, `1`
configuration/dependency/operation failure. `migrate` and `provision` require
`--confirm`; help and argument parsing do not open a database connection.

## Deferred acceptance checklist

None of these boxes is closed by a build or a skipped test:

- [ ] Approve an isolated loopback-only disposable database and private
  short-lived credential references; confirm database name/role/host and cleanup.
- [ ] Execute migrations on a fresh database; rerun for idempotence; exercise
  checksum failure, lock contention, deadline and rollback/restore privately.
- [ ] Verify provisioning transaction/FKs/audit/access, duplicate-identity
  rollback, exact issuer matching and interrupted-commit reconciliation.
- [ ] Confirm `db-ping` performs no schema writes and `serve` performs no
  migrations or provisioning; missing schema must not be mistaken for ready.
- [ ] Approve a real development Authentik client, exact callback, RS256/JWKS
  and HTTPS proxy. Exercise discovery failure/mismatch and expired/rotated keys.
- [ ] Complete real OIDC login/logout, nonce/state/PKCE/CSRF failures, cookie
  flags, unmapped/inactive identity rejection and session expiry/revocation.
- [ ] Verify `/api/healthz` versus dependency `/api/readyz`, frontend assets,
  authorization isolation and task routes against the built app.
- [ ] Send SIGINT/SIGTERM during active requests; verify completion/deadline,
  listener closure, pool cleanup and nonzero bind/dependency failure behavior.
- [ ] Inspect private logs for credential/DSN/token/code leakage and verify only
  fixed safe startup/admin errors escape; do not publish those private logs.
- [ ] Run the separately maintained full module/integration/browser/load/restore
  suites after their approved prerequisites exist. Record actual execution,
  blockers and cleanup; a deferred checklist is not a successful test report.
