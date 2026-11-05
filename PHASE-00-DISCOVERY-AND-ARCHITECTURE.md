# Phase 0 — Discovery and Architecture

**Project:** NEWCONNEXT — AI Marketing Office  
**Phase type:** Discovery, architecture, and technical validation  
**Status:** Must be completed before major implementation

## 1. Objective

Establish a verified technical blueprint for NEWCONNEXT before building application features. Discover the existing environment, preserve the six existing Hermes Agent profiles, validate integration capabilities, and document architectural decisions.

Do not begin extensive feature implementation during this phase.

## 2. Scope

### 2.1 Environment discovery
Inspect and document:

- Existing source repositories and branches.
- VPS operating system, CPU, memory, disk, and network constraints.
- Existing Docker and Docker Compose versions.
- Existing reverse proxy and domain configuration.
- Existing PostgreSQL, Redis, MinIO, Temporal, Authentik, Postiz, and Superset deployments, if any.
- Current backup and recovery arrangements.
- Existing firewall and exposed ports.
- Existing CI/CD and source-control arrangements.

Do not expose secrets in reports. Record only whether a required secret/configuration exists, not its value.

### 2.2 Hermes discovery
For all six existing agents, verify:

- Display name and role.
- Telegram username.
- Actual Hermes identifier, if available.
- Supported invocation mechanism.
- Whether task execution can be initiated programmatically.
- Whether execution status can be retrieved.
- Whether results can be delivered through callbacks, polling, or another supported mechanism.
- Whether cancellation, timeout, and retry are supported.
- Whether group messages and bot privacy settings affect communication.

Do not infer that Telegram bot access implies an API or reliable agent-to-agent communication. Do not modify agent personalities or system prompts.

### 2.3 Postiz discovery
Verify against current official documentation and the deployed version:

- Authentication mechanism.
- API availability and version.
- Channel/account management.
- Media upload requirements.
- Scheduling and publishing endpoints.
- Status and error retrieval.
- Webhook support.
- Rate limits and platform-specific restrictions.
- Supported networks and content formats.

Record unsupported or unverified capabilities explicitly.

### 2.4 Identity and security discovery
Validate the proposed Authentik OIDC design:

- OIDC issuer and discovery endpoint.
- Authorization Code Flow with PKCE support.
- Client configuration requirements.
- Redirect URI requirements.
- Claims required for user mapping.
- MFA and passkey capabilities.
- Session and logout behavior.
- Reverse-proxy requirements.

### 2.5 Architecture discovery
Confirm the feasibility of:

- Go modular monolith using Hexagonal Architecture.
- React + TypeScript frontend.
- PostgreSQL as business system of record.
- MinIO for files and media.
- Temporal for durable workflows.
- Redis for cache/transient coordination where justified.
- Postiz as publishing adapter.
- Superset for embedded analytics.
- Docker Compose deployment.

## 3. Required architecture principles

- Use Hexagonal Architecture (Ports and Adapters).
- Keep domain logic independent from frameworks and external services.
- Use application use cases to coordinate business operations.
- Keep HTTP handlers thin.
- Isolate external systems behind adapters.
- Use PostgreSQL as the authoritative business data store.
- Use Temporal for workflow execution, not as a replacement for business records.
- Use an outbox or equivalent reliable event-delivery pattern for critical asynchronous operations.
- Avoid premature microservices.
- Preserve all six existing Hermes profiles.

## 4. Required deliverables

Create and maintain:

- `docs/00-project-overview.md`
- `docs/02-architecture-overview.md`
- `docs/03-hexagonal-architecture.md`
- `docs/04-project-structure.md`
- `docs/05-domain-model.md`
- `docs/06-database-design.md`
- `docs/08-authentication-security.md`
- `docs/09-hermes-integration.md`
- `docs/10-postiz-integration.md`
- `docs/architecture/system-context.md`
- `docs/architecture/container-diagram.md`
- `docs/architecture/component-diagram.md`
- `docs/architecture/deployment-diagram.md`
- `docs/architecture/data-flow.md`
- `docs/adr/ADR-0001-initial-architecture.md`
- `docs/adr/ADR-0002-human-authentication.md`
- `docs/adr/ADR-0003-agent-integration.md`
- `docs/adr/ADR-0004-workflow-orchestration.md`

Include Mermaid diagrams for system context, containers, components, deployment, and key workflows.

Produce an initial ERD showing organizations, users, roles, brands, agents, tasks, campaigns, content, assets, approvals, integrations, and audit records.

## 5. Acceptance criteria

Phase 0 is complete only when:

- Existing environment has been inspected.
- Hermes invocation and result-delivery mechanisms are verified or clearly marked as unresolved.
- Postiz capabilities are verified against the actual version/API.
- OIDC feasibility is documented.
- Security and deployment constraints are recorded.
- Architecture and data-flow diagrams exist.
- Initial ERD exists.
- ADRs explain significant choices and alternatives.
- Risks, assumptions, dependencies, and open questions are listed.
- A phase-by-phase implementation plan is presented.

## 6. Phase report format

At completion, report:

1. Environment discovered.
2. Verified integration capabilities.
3. Unverified capabilities and blockers.
4. Proposed architecture.
5. Mermaid diagrams and ERD locations.
6. ADR list.
7. Security risks.
8. Infrastructure risks.
9. Recommended implementation order.
10. Explicit approval requests for major infrastructure changes.

## 7. Out of scope

- Building all application modules.
- Rewriting Hermes agent profiles.
- Publishing real social media content.
- Changing production credentials.
- Destructive infrastructure changes.

## 8. Handoff to Phase 1

Phase 1 may begin after the architecture blueprint and implementation plan have been reviewed and approved.
