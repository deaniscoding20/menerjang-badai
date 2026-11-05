# Phase 1 — Application Foundation, Identity, and Security

**Prerequisite:** Approved Phase 0 architecture  
**Primary outcome:** A secure, runnable application foundation

## 1. Objective

Create the NEWCONNEXT application skeleton with Go backend, React frontend, PostgreSQL, migrations, OIDC human authentication, RBAC foundations, audit logging, and Docker Compose development environment.

Implement Hexagonal Architecture from the first commit.

## 2. Scope

### 2.1 Repository and project structure
Create the agreed repository layout for:

- Go API.
- Go workflow worker entry point (worker may initially be a skeleton).
- React frontend.
- Database migrations.
- OpenAPI specification.
- Docker deployment.
- Markdown documentation.
- Tests and scripts.

Do not create empty abstractions or packages without a current responsibility.

### 2.2 Backend foundation
Implement:

- Application bootstrap and configuration loading.
- HTTP server using the selected router.
- Graceful shutdown.
- Request IDs and correlation IDs.
- Structured logging.
- Centralized error mapping.
- Health and readiness endpoints.
- Dependency wiring through constructors.
- OpenAPI generation or maintained OpenAPI contract.

Use explicit interfaces at application boundaries.

### 2.3 Database foundation
Implement:

- PostgreSQL connection management.
- Migration tooling.
- sqlc configuration and generated query workflow.
- Transaction helper or explicit transaction boundaries.
- Separate least-privilege application and migration database roles where appropriate.

Initial schema should cover:

- organizations
- users
- roles
- permissions
- user_roles
- brands (minimal foundation)
- user_brand_access
- sessions, if server-managed application sessions are used
- audit_logs

Use UUIDs where appropriate, UTC timestamps, constraints, and indexes.

### 2.4 Human authentication
Integrate Authentik using OIDC.

Required flow:

1. User selects login.
2. Backend creates OIDC authorization request with state, nonce, and PKCE.
3. User authenticates at Authentik.
4. Backend validates callback, state, nonce, issuer, audience, signature, and claims.
5. Backend maps external identity to an application user.
6. Backend creates a secure application session.
7. Browser receives an opaque HttpOnly cookie.

Use Authorization Code Flow with PKCE (S256).

Do not expose client secrets or long-lived tokens to frontend JavaScript.

Implement logout and server-side session revocation. Document provider logout behavior and any limitations.

### 2.5 Session security
Implement:

- Secure cookie attributes in production.
- HttpOnly.
- SameSite=Lax or stricter where compatible.
- CSRF protection for state-changing requests.
- Session rotation after login and privilege changes.
- Idle and absolute expiration.
- Session revocation.
- Reauthentication policy for sensitive operations.

Do not store long-lived tokens in localStorage.

### 2.6 RBAC foundation
Create roles and permission checks on the backend.

Initial roles:

- Owner
- Administrator
- Marketing Manager
- Editor
- Approver
- Analyst
- Viewer

Implement permission middleware or application-level authorization policies, while ensuring business use cases also enforce critical authorization rules.

Never treat frontend visibility as authorization.

### 2.7 Frontend foundation
Implement:

- Login page.
- Authenticated application shell.
- Navigation layout.
- User profile menu.
- Session-expired handling.
- Access-denied page.
- Loading, empty, and error states.
- Responsive layout.

Use TypeScript strict mode, reusable components, accessible controls, and consistent form validation.

### 2.8 Audit logging
Record security-relevant events:

- Login success/failure where available.
- Logout.
- Session revocation.
- User invitation and role changes.
- Permission changes.
- Administrative configuration changes.

Never log secrets, tokens, passwords, or session cookies.

## 3. Hexagonal boundaries

Domain must not import OIDC, HTTP, SQL, or framework packages.

Application defines identity and authorization ports.

OIDC and session implementations live in outbound adapters.

HTTP routes and handlers live in inbound adapters.

Database implementations live in PostgreSQL adapters.

## 4. Required deliverables

- Runnable Go API.
- Runnable React application.
- PostgreSQL migrations.
- Authentik OIDC configuration guide.
- RBAC seed strategy.
- Audit log foundation.
- Docker Compose development environment.
- `.env.example` with placeholders only.
- OpenAPI documentation.
- Unit and integration tests.
- `docs/08-authentication-security.md`
- `docs/14-local-development.md`
- Updated architecture and project structure documentation.

## 5. Acceptance criteria

- A permitted human user can log in through Authentik.
- Invalid OIDC callbacks are rejected.
- CSRF protections are tested.
- User session can be revoked.
- Unauthorized backend requests are rejected.
- A user cannot access a brand without brand permission.
- No secret is committed to the repository.
- Migrations work from an empty database.
- API and frontend start through documented commands.
- Health and readiness endpoints behave correctly.
- Tests and linters pass.

## 6. Handoff to Phase 2

Provide stable identity, organization, authorization, and audit foundations before adding agent and brand operations.
