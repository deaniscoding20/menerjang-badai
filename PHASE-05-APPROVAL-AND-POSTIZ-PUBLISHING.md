# Phase 5 — Human Approval Center and Postiz Publishing

**Prerequisite:** Phase 4 acceptance criteria passed  
**Primary outcome:** Controlled, auditable publishing of approved content

## 1. Objective

Implement human review and approval, then integrate Postiz for scheduling and publishing through verified capabilities.

Publishing must never occur merely because an agent completed a task.

## 2. Approval Center

Support:

- Submit for review.
- Approve.
- Reject.
- Request revision.
- Comments.
- Version comparison.
- Asset preview.
- Approval history.
- Assigned approver.
- Review deadline.
- Notifications.

Approval record must bind to:

- Exact content version.
- Exact asset versions.
- Brand.
- Intended platform/channel.
- Publication intent and schedule where applicable.
- Approver.
- Decision timestamp.
- Comments.

## 3. Approval invalidation

If approved content, media, destination channel, or other material publishing details change, the approval must be invalidated or a new approval required.

Do not allow approval to silently transfer to a materially different content version.

Enforce this rule in the backend, not only in the UI.

## 4. Human authorization

Only authorized users may approve and publish.

Require explicit human authorization for:

- Publishing.
- Changes to connected social accounts.
- Advertising spend.
- Brand identity changes.
- Destructive external operations.

Consider step-up authentication for high-impact actions.

## 5. Postiz adapter

Verify the deployed Postiz version and official API documentation before coding.

Implement only verified capabilities.

Document:

- Authentication.
- Channel mapping.
- Media upload flow.
- Scheduling.
- Publishing.
- Status synchronization.
- Error mapping.
- Rate limits.
- Webhook behavior.
- Supported networks and formats.

Do not invent endpoints or assume every social network is supported.

## 6. Source-of-truth boundaries

PostgreSQL is authoritative for:

- Content.
- Content versions.
- Approval decisions.
- Campaigns.
- Intended publication.
- Internal publication state.
- Audit history.

Postiz is the external scheduling/publishing engine.

Store external Postiz IDs and synchronization metadata, not duplicate business truth.

## 7. Publishing workflow

1. Content version is submitted for review.
2. Approver reviews content and assets.
3. Approver approves or requests changes.
4. Backend verifies approval is still valid.
5. Authorized user confirms publication intent and schedule.
6. Backend creates an idempotent publishing job.
7. Postiz adapter submits the operation.
8. Application synchronizes external status.
9. Errors are recorded and surfaced.
10. User can inspect the complete audit trail.

## 8. Reliability

Implement:

- Idempotency keys.
- Duplicate prevention.
- Retry classification.
- Exponential backoff where appropriate.
- Dead-letter or failed-operation visibility.
- Webhook signature verification.
- Safe polling fallback if webhooks are unavailable.
- Reconciliation for uncertain outcomes.

Never blindly retry an operation if the external publishing outcome is unknown. Reconcile external state first to avoid duplicate posts.

## 9. User interface

Implement:

- Approval inbox.
- Review detail page.
- Content and asset preview.
- Revision comparison.
- Approval actions.
- Publishing confirmation.
- Scheduling interface.
- Publishing job status.
- Failed publishing view.
- Audit history.

## 10. Acceptance criteria

- Unapproved content cannot be published through application APIs.
- Approval references exact content and asset versions.
- Material changes invalidate approval.
- Only authorized users can approve/publish.
- Postiz integration uses verified endpoints.
- Duplicate publishing is prevented or reconciled safely.
- Publishing failures are visible and actionable.
- External IDs and statuses are synchronized.
- All actions are auditable.

## 11. Handoff to Phase 6

Provide trustworthy publishing records and channel mappings for analytics ingestion and performance reporting.
