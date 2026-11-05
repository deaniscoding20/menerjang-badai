# Phase 3 — Hermes Integration, Task Management, and Workflows

**Prerequisite:** Phase 2 acceptance criteria passed  
**Primary outcome:** Dispatch tasks to Hermes and track durable work

## 1. Objective

Connect the application to the six existing Hermes Agents through verified integration mechanisms. Build task management and durable orchestration without making Hermes the source of truth for business state.

## 2. Integration discovery gate

Before implementation, confirm the actual Hermes communication interface.

Possible mechanisms may include a supported API, webhook, message gateway, or documented invocation mechanism.

Do not invent endpoints or assume Telegram group messaging is sufficient.

For each agent, document:

- How a task is submitted.
- How execution is correlated with an application task.
- How completion is detected.
- How output is retrieved.
- How errors are reported.
- Whether cancellation is supported.
- Whether retries are safe.
- Authentication and network requirements.

If a capability is unavailable, document the limitation and implement only the verified subset.

## 3. Task domain

Implement task lifecycle:

- Draft.
- Ready.
- Assigned.
- Queued.
- Running.
- Waiting for input.
- Waiting for approval.
- Completed.
- Failed.
- Cancelled.

A task should include:

- Task ID.
- Organization and brand.
- Requester.
- Assignee (human or agent).
- Task type.
- Objective.
- Instructions.
- Priority.
- Deadline.
- Dependencies.
- Input references.
- Expected deliverable schema.
- Current status.
- Execution correlation ID.
- Created/updated timestamps.

Keep task events and status history.

## 4. Task management features

Implement:

- Create and edit tasks.
- Assign or reassign.
- Direct assignment to any specialist.
- Task dependencies.
- Priority and deadline.
- Comments and clarifications.
- Deliverable submission.
- Revision request.
- Cancellation.
- Task history.
- Agent execution logs.

Do not force all tasks through Claudia.

## 5. Structured agent requests and deliverables

Define versioned task request and response schemas.

Support structured deliverables such as:

- Research report.
- Marketing strategy.
- Creative brief.
- Caption package.
- Design specification.
- Campaign plan.
- Performance report.

Store raw agent output where appropriate and a validated structured representation.

Mark unvalidated or malformed output clearly. Never silently treat arbitrary text as a valid structured deliverable.

## 6. Temporal integration

Use Temporal for durable long-running workflows.

Implement:

- Workflow start.
- Activity execution.
- Retry policies.
- Timeouts.
- Heartbeats where applicable.
- Human approval pauses.
- Workflow cancellation.
- Failure visibility.
- Recovery after process restarts.

PostgreSQL remains the authoritative business record.

Temporal owns workflow execution state and history.

## 7. Reliable asynchronous processing

Use a transactional outbox or equivalent reliable dispatch pattern.

Requirements:

- Persist business state before dispatch.
- Ensure events are delivered at least once where applicable.
- Make consumers idempotent.
- Use idempotency keys for external side effects.
- Prevent duplicate task dispatch where possible.
- Record dispatch attempts and outcomes.
- Handle delayed and out-of-order callbacks safely.

Do not claim exactly-once delivery unless the complete system can guarantee it.

## 8. Workflow examples

### Marketing strategy workflow
1. Human creates strategy task.
2. Jordan is assigned.
3. Curry research may be requested as a dependency or parallel task.
4. Results are attached to the task.
5. Jordan submits strategy.
6. Human reviews or requests revision.

### Campaign coordination workflow
1. Claudia organizes the brief.
2. Jordan and Curry work on strategy and research.
3. Kobe develops creative direction.
4. Sabrina and Kyrie work in parallel.
5. Kobe performs creative QA.
6. Claudia consolidates.
7. Human approves.

Direct specialist assignments remain supported.

## 9. User interface

Implement:

- Task list.
- Kanban board.
- Task detail.
- Assignment controls.
- Dependency visualization.
- Agent execution status.
- Deliverable viewer.
- Workflow detail and timeline.
- Retry/cancel actions with permission checks.

## 10. Acceptance criteria

- A task can be created and assigned to any configured agent.
- Dispatch uses a verified Hermes mechanism.
- Execution can be correlated with the application task.
- Deliverables are stored and versioned.
- Failures are visible.
- Safe retries do not create unintended duplicate work.
- Workflow state recovers after worker restart.
- Direct assignment bypasses Claudia where requested.
- All important actions are auditable.

## 11. Handoff to Phase 4

Provide reliable task, workflow, deliverable, and agent execution foundations for knowledge management and content production.
