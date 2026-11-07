# Hermes adapter — bounded discovery and candidate contract

## Status

**Development discovery only. Dispatch is still disabled.** This document records read-only evidence and an implementation boundary; it does not certify gateway isolation, caller authorization, task execution, retry safety, or durable workflows.

Official reference re-read: <https://hermes-agent.nousresearch.com/docs/user-guide/messaging/a2a>. Installed source inspected: `plugins/platforms/a2a/{adapter,protocol}.py`. The source hashes below match the earlier [discovery](09-hermes-integration.md); disk identity does not establish what a live process imported.

## Read-only observation

At `2026-11-06T07:15:10Z`, GET of six loopback Agent Cards returned HTTP200 and each advertised `supportedInterfaces[].protocolVersion: "1.0"`, `protocolBinding: "JSONRPC"`. Protocol version is **inside supportedInterfaces**, not a top-level `protocolVersion` property. The separate card `version` was `1.0.0`.

- Claudia: `/claudia/`, name Claudia, streaming advertised true.
- Jordan: `/jordan/`, name Jordan, streaming advertised false.
- Kobe: `/kobe/`, name Kobe, streaming advertised false.
- Curry: `/curry/`, name Curry, streaming advertised false.
- Sabrina: `/sabrina/`, name Sabrina, streaming advertised false.
- Kyrie: `/kyrie/`, name Kyrie, streaming advertised false.

All routes are beneath `http://127.0.0.1:9900`. Cards advertise push notifications, but delivery, signatures, replay protection and receiver reachability have not been exercised. Capability advertisements are not behavioral acceptance. No POST, credential read, model call, profile/config modification or restart occurred during this discovery.

Socket observation: listener `127.0.0.1:9900`, Hermes PID256161. This establishes the observed address/PID, not isolation or import provenance. Private gitignored evidence: `artifacts/testing/hermes-card-discovery-20261106T071510Z/discovery.json`; run0700/report0600. All six route results were saved, deduplicated and counted in code:6 unique routes,6 HTTP200/protocol1.0 cards.

Source SHA256:

```text
adapter.py  12a8bc5416e12747f2687a25eeaac27565287c916363f3e4b8be525d4af7ea42
protocol.py 099f0a25cb5629e3e64bb3d45ea3f68894594427a74b52a5473cf2073b760924
tools.py    9a08c4fbf49c00dbb950f8179d7b2eb33b10d882d405969c2dc5ffa1246b9d0d
```

## Candidate outbound contract — source-inspected, not live wire capture

The application must select an exact profile route from the server registry. Never use a URL, tenant, bearer token or callback supplied by the browser or deliverable. Application organization/brand scope is distinct from the Hermes profile tenant.

1. Use JSON-RPC2.0 method `SendMessage` with `params.tenant` derived from the route and `params.message` containing `messageId`, `contextId`, `role:"ROLE_USER"`, and bounded text `parts`.
2. Persist immutable application attempt identity, request schema version/payload digest, organization/brand/task, destination and stable wire context/message IDs **before** any dispatch, through an atomic business-state/outbox transaction.
3. Canonical SendMessage result is a oneof `{task: ...}` or `{message: ...}`. `protocol.send_message_response` wraps an object with id/status as task; a legacy bare task also exists for path-style aliases. Do not accidentally parse the canonical wrapper as a bare task. A future initial adapter should pin the canonical shape and fail safely on an unsupported alternative rather than invent business success.
4. A task contains `id`, `contextId`, `status.state`, `status.timestamp`, optional `status.message` and optional artifacts. Completed output is emitted as text artifact parts; output can also occur in a status message. Avoid concatenating duplicated artifact/status copies. Nontext artifacts and URLs are untrusted references, not instructions to download or execute.
5. Recognized source states are `TASK_STATE_SUBMITTED`, `TASK_STATE_WORKING`, `TASK_STATE_INPUT_REQUIRED`, `TASK_STATE_COMPLETED`, `TASK_STATE_FAILED`, `TASK_STATE_CANCELED`, and `TASK_STATE_REJECTED`. Unknown/malformed states must remain visibly unsupported, never completed.
6. `GetTask` accepts source-inspected `taskId` or `id`; both lookup and cancellation apply exact route/tenant and authenticated peer ownership. A successful transport completion is not approval, schema validation, factual correctness, publication readiness, or application task completion.
7. Raw bounded output and a separately validated structured envelope must be stored with server-derived authenticated attempt/agent provenance. Never repurpose the human-import endpoint by allowing an `origin` or `agent_id` body field. Do not persist bearer tokens, raw provider errors, credentials or secret-bearing URLs in audit/telemetry.

These are candidate codec/adapter requirements, not evidence of a configured application service caller. The current application has no live Hermes dispatch adapter or enabled worker.

## Retry and cancellation limitations discovered in source

- `adapter.py:595–619` reads `messageId` for **forwarded routes only**. The local route does not establish equivalent stable-message deduplication. In this deployment the cards distinguish the local Claudia route from forwarded specialist routes; direct Claudia retries cannot be assumed safe.
- Forwarded owner admission keeps an ID derived from exact context/message, while adapter retry bookkeeping is separately scoped. That is source behavior, not an end-to-end durability guarantee across upgrades, process loss or downstream tools.
- A SendMessage request can synchronously wait for the reply. A disconnect/timeout may occur after admission. `_await_reply` can produce a failed protocol state for reply timeout; that alone is not proof work never ran or has stopped. Preserve **unknown external outcome**, and reconcile before operator-authorized resubmission.
- `CancelTask` updates protocol task state/resolves a pending wait. It is not demonstrated to stop an owner/model run or external side effects. Record application cancellation separately from the external acknowledgment.
- TaskStore is in-memory. Agent Card availability and persisted conversations do not prove task polling/subscriptions/cancellation survive gateway restart. PostgreSQL remains business authority; Temporal must own durable orchestration.

Consequently: no automatic resend for either local or forwarded routes until route-specific admission and recovery are verified. No exactly-once claim. A new revision request is a new human business record, not permission to replay an old external attempt.

## Next gates

Before enabling any execution, provide an explicit least-privilege service caller without copying peer credentials into Git/browser, verify active process/source and route ownership with synthetic authorized evidence, agree host/worker private networking, and implement transactional outbox, attempt records, reconciliation and Temporal lifecycle. Keep failure/restore/restart/load campaigns in the deferred [testing register](testing/DEFERRED-CHECKLIST.md), not against existing production profiles by default.

Fast development can implement pure request/response codecs and synthetic transport tests against this inspected contract. Such fixtures must be labeled source-derived and cannot close live authorization, execution, deduplication or restart gates. No production cutover is implied by this document.
