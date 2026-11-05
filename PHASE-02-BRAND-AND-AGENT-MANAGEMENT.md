# Phase 2 — Brand and Agent Management

**Prerequisite:** Phase 1 acceptance criteria passed  
**Primary outcome:** Manage brands, users' brand access, and the six existing Hermes Agents

## 1. Objective

Implement multi-brand management and an Agent Registry. The registry represents the six existing Hermes Agents without changing their profiles or making assumptions about their internal identifiers.

## 2. Brand management

Implement:

- Create, view, update, archive brands.
- Brand profile.
- Logo and basic identity metadata.
- Website and business description.
- Products and services.
- Target audience.
- Positioning and value proposition fields.
- Tone-of-voice summary.
- Brand status.
- Brand owner and authorized members.

Enforce brand-level access on every read and write.

Do not rely solely on filtering data in the frontend.

## 3. Brand data isolation

Define and test the authorization model:

- Organization-level membership.
- Brand-level membership.
- Role and permission inheritance rules.
- Explicit access denial when a user has no brand permission.
- Cross-brand query protections.

Document how background workers obtain authorized brand context.

## 4. Agent Registry

Seed or configure the six existing agent records:

| Agent | Role | Telegram |
|---|---|---|
| Claudia | Assistant Brand Manager & Coordinator | `@MB_claudia_bot` |
| Jordan | Marketing Strategist | `@MB_jordan_bot` |
| Kobe | Creative Lead | `@MB_kobe_bot` |
| Curry | Market Researcher | `@MB_curry_bot` |
| Sabrina | Social Media & Copywriter | `@MB_sabrina_bot` |
| Kyrie | Creative Designer | `@MB_kyrie_bot` |

Each record should support:

- Stable application agent ID.
- Display name.
- Role and responsibilities.
- Telegram username.
- External Hermes identifier, if discovered.
- Integration type.
- Capabilities.
- Enabled/disabled status.
- Connection status.
- Last observed activity.
- Created and updated timestamps.

Application agent IDs must not depend on Telegram usernames or Hermes internal IDs.

Do not modify the agents' prompts, personalities, or existing configuration.

## 5. Agent capabilities and status

Represent capabilities as structured metadata, not free-form assumptions.

Distinguish:

- Configured.
- Reachable.
- Ready to accept work.
- Busy.
- Unavailable.
- Error.
- Unknown.

Do not show an agent as online merely because its Telegram bot exists.

Status must be based on a verified heartbeat, API response, or other documented signal.

## 6. User interface

Implement:

- Brand list and detail pages.
- Brand member management.
- Agent directory.
- Agent detail page.
- Organization overview.
- Agent role and collaboration visualization.
- Integration status indicators.
- Empty and error states.

The organization diagram must show responsibilities, not imply that every task has to pass through Claudia.

## 7. Database

Implement and document appropriate entities, including:

- brands
- brand_members or user_brand_access
- brand_profiles
- agents
- agent_capabilities
- agent_connections

Use migrations and sqlc-generated query code.

## 8. Tests

Test:

- Brand CRUD authorization.
- Cross-brand access denial.
- User invitation or membership assignment.
- Agent identity mapping.
- Agent status interpretation.
- Seed idempotency.
- Disabled agent behavior.

## 9. Acceptance criteria

- All six agents appear in the dashboard.
- Their names, roles, and Telegram usernames are correct.
- Existing agent profiles are not overwritten.
- Application agent IDs are stable and independent.
- Brand access is enforced by the backend.
- Agent status is not fabricated.
- Administrators can manage brands and membership.
- Documentation describes the brand and agent domain models.

## 10. Handoff to Phase 3

Provide stable agent identities, capability metadata, brand authorization, and registry APIs for task dispatch and workflow orchestration.
