# Phase 7 — Production Hardening, Security, and Operations

**Prerequisite:** Phases 1–6 acceptance criteria passed  
**Primary outcome:** A supportable, secure, observable, recoverable production system

## 1. Objective

Review the complete platform for security, reliability, performance, operational readiness, and maintainability.

Do not declare production readiness based only on a successful local startup.

## 2. Security review

Review:

- OIDC configuration.
- Session lifecycle.
- MFA enforcement.
- CSRF protection.
- CORS policy.
- Cookie attributes.
- RBAC and brand isolation.
- API authorization.
- Webhook authentication.
- Secret storage.
- File upload validation.
- SQL query safety.
- XSS protection.
- Rate limiting.
- Audit logging.
- Container privileges.
- Network exposure.
- Dependency vulnerabilities.

Test cross-brand access attempts and privilege escalation scenarios.

Do not log credentials, tokens, cookies, or sensitive payloads.

## 3. Infrastructure hardening

Review Docker Compose production configuration:

- Non-root containers where feasible.
- Minimal exposed ports.
- Internal networks.
- Persistent volumes.
- Health checks.
- Restart policies.
- Resource limits.
- Separate credentials per service.
- Secret injection.
- Image pinning.
- Container image scanning.
- TLS through Caddy.
- Administrative interfaces restricted to trusted access.

Do not expose PostgreSQL, Redis, MinIO administration, Temporal administration, or Superset administration publicly without a documented protected access path.

## 4. Backup and recovery

Create and test:

- PostgreSQL backup.
- PostgreSQL restore.
- MinIO backup.
- MinIO restore.
- Configuration backup.
- Encryption key recovery procedure.
- Recovery ordering.
- Retention policy.
- Backup integrity verification.

Document recovery point objective (RPO) and recovery time objective (RTO) as explicit targets agreed with the owner.

A backup is not considered valid until restoration has been tested.

## 5. Observability

Implement and verify:

- Structured logs.
- Correlation IDs.
- Request metrics.
- Error metrics.
- Latency metrics.
- Database health.
- Worker health.
- Temporal workflow health.
- Hermes integration health.
- Postiz integration health.
- MinIO health.
- Social data ingestion health.
- AI usage and cost metrics.

Use Prometheus, Grafana, and OpenTelemetry as appropriate.

Ensure sensitive data is not included in telemetry.

## 6. Reliability testing

Test:

- API restart.
- Worker restart.
- Database connection interruption.
- Redis interruption where used.
- MinIO interruption.
- Temporal worker interruption.
- Hermes timeout.
- Hermes malformed response.
- Postiz timeout and uncertain publishing outcome.
- Duplicate webhook delivery.
- Delayed and out-of-order events.
- Failed social data sync.

Verify recovery and reconciliation behavior.

## 7. Performance testing

Establish realistic expected workloads.

Measure:

- API latency.
- Database query performance.
- Concurrent user behavior.
- File upload performance.
- Workflow throughput.
- Agent dispatch latency.
- Analytics query performance.

Use load testing only with safe test data and non-production external integrations.

Set measurable performance targets rather than making unsupported claims.

## 8. Operational runbooks

Create:

- Backup and restore.
- Incident response.
- Credential rotation.
- OIDC outage.
- Hermes integration failure.
- Postiz publishing failure.
- MinIO storage failure.
- PostgreSQL recovery.
- Temporal workflow recovery.
- Social API rate-limit handling.
- Deployment rollback.
- Upgrade procedure.

## 9. CI/CD readiness

Provide pipeline steps for:

- Formatting.
- Linting.
- Unit tests.
- Integration tests.
- Frontend checks.
- Dependency scanning.
- Container build.
- Image scanning.
- Migration validation.
- Deployment approval.

Production deployment must require an explicit release process.

## 10. Documentation audit

Verify that:

- README matches actual commands.
- Architecture diagrams match implementation.
- OpenAPI matches API behavior.
- Database documentation matches migrations.
- Environment variable documentation is complete.
- Runbooks have been tested.
- Known limitations are visible.
- No sample secret resembles a real production credential.

## 11. Acceptance criteria

- Critical security tests pass.
- Cross-brand isolation tests pass.
- Backup restoration has been demonstrated.
- Operational dashboards work.
- Alerts are configured for critical failures.
- Deployment and rollback procedures are documented.
- Integration failure scenarios have been tested.
- No unresolved critical security findings remain.
- Production readiness report is approved by the owner.

## 12. Final handoff

Provide:

1. Production architecture.
2. Deployment instructions.
3. Environment variable reference.
4. Backup and restore instructions.
5. Security review summary.
6. Test and load-test results.
7. Monitoring dashboards.
8. Operational runbooks.
9. Known limitations.
10. Recommended next development priorities.
