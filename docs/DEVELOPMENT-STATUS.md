# Development snapshot — Phase 1–3 work in progress

This branch publishes the current source so the owner can inspect progress. **It is not a completed Phase 3, a release, a deployment, or proof of live integration.**

## Implemented source

- Go modular application, OIDC/session and CSRF foundation, organization/brand access controls, React/TypeScript workspace and six-agent registry.
- Task list/board/detail, draft editing, priority/deadline, direct specialist assignment metadata, manual preparation transitions and history.
- Human comments/clarifications, same-brand dependency graph, immutable versioned human-imported deliverables and immutable revision-request notes.
- PostgreSQL migrations `000001`–`000006`, repository implementations, application/HTTP tests, frontend fixture tests, local startup commands and deferred testing setup.
- Production readiness remains false. Runtime task transitions that would imply agent execution remain disabled.

## Checks rerun for this snapshot

Checked on 2026-11-07 UTC:

- Frontend: **260 tests across 15 files passed**; typecheck, lint and production build passed. Two nonfatal Zod/Rollup comment-annotation warnings remain.
- Go: domain, application, HTTP, CLI and platform short tests passed; all-package build and scoped vet passed.
- PostgreSQL adapter and acceptance fixtures: **compile-only passed**, with no database tests executed. The new late-event/audit rollback fixture was formatted and compiled, not run against PostgreSQL.
- Python policy/setup suite: **38 tests run, 35 passed and 3 errors**. All three are existing `tests/test_vps_delivery.py` filesystem cases referencing missing `scripts.vps_delivery.trusted_json`; this snapshot does not fix or hide them.
- Local documentation links/targeted secret checks, Gitleaks directory scan and Actions lint passed. Release marker validation reports `production deployable: false`.

These checks do not constitute complete CI, integration, E2E, security, load or release acceptance.

## Reviewed UI race fix

The parent-transition-first race is fixed in source with a synchronous mutation ref checked by transitions, shared panel admission and brief editing. The regression first failed with two POSTs, then passed with exactly one transition POST using coherent version-7 snapshots, blocking through the follow-up refresh and preserving the unsent revision note, explicit historical target and raw import draft. The lock releases in `finally`; the existing task-refresh guard remains active. All 260 frontend tests, typecheck, lint and build passed; a later focused 49-test rerun also passed.

Independent bounded static review `deleg_d06bdc29` returned `passed=true` with no security/logic findings or suggestions. Persistence review also found no blocking issues. Neither review establishes live database or browser acceptance.

## Phase 3 remaining

1. Reciprocal parent-transition/panel-mutation race: source fix and focused regression complete; live browser acceptance remains deferred.
2. Implement versioned execution-request envelopes and strict source-derived Hermes response/transport contracts. Pure request-contract discovery is complete; implementation is not.
3. Add authoritative attempt/correlation records, transactional outbox, idempotent consumers and outcome reconciliation.
4. Implement Temporal workflow/activity lifecycle, bounded retry/timeout policies, human input/approval pauses, cancellation semantics, execution status/logs and workflow timeline.
5. Implement authenticated result ingestion with immutable versioned raw/validated output. A completed transport is not human approval or business completion.
6. Satisfy the existing Hermes caller authorization, route isolation and active-runtime discovery gates before enabling traffic; retain unknown-outcome/no-blind-resend behavior.
7. Run the agreed deferred integration/restart/recovery/E2E campaigns and close acceptance gates with real evidence. They remain pending, not passed.

## Preserved operational boundaries

No production migration, runtime Hermes/profile change, model dispatch, website replacement, merge to main or deployment is part of this publication. Existing site, Hermes dashboard, personas, permissions and credentials are not changed by pushing this feature branch. No secrets, session databases, transcripts, build output or local testing evidence are intended for publication.

See [task implementation details](12-task-development.md), [local startup](14-local-development.md), [Hermes adapter contract](15-hermes-adapter-contract.md), [deferred cases](testing/DEFERRED-CHECKLIST.md) and [testing setup](../testing/README.md).
