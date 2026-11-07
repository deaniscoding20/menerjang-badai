package postgres

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/menerjangbadai/dashboard/internal/application"
	"github.com/menerjangbadai/dashboard/internal/domain"
)

var _ application.TaskRevisionRequestRepository = (*Store)(nil)

// Deferred acceptance only: testStore requires an explicit disposable database.
// Compile with MB_INTEGRATION=0 MB_TEST_DATABASE_URL= go test ./internal/adapters/postgres -run '^$'.
func TestRealRevisionRequestsBindImmutableVersionAndAudit(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()
	actor, err := store.Bootstrap(ctx, "https://issuer.test/", uuid.NewString(), "Revision owner")
	if err != nil {
		t.Fatal(err)
	}
	brand, err := (application.Brands{Repo: store}).Create(ctx, actor, domain.Brand{Name: "Revision fixture"})
	if err != nil {
		t.Fatal(err)
	}
	service := application.Tasks{Repo: store}
	task, err := service.Create(ctx, actor, domain.Task{BrandID: brand.ID, Title: "Review research"})
	if err != nil {
		t.Fatal(err)
	}
	raw := ` {"schema_version":1,"kind":"research_report","title":"Audience","sections":[{"heading":"Evidence","body":"Original finding"}]} `
	first, err := store.SubmitTaskDeliverable(ctx, actor, task.ID, domain.DeliverableSubmission{RawOutput: raw, Version: 1})
	if err != nil || first.ValidationStatus != "validated" {
		t.Fatal(first, err)
	}
	second, err := store.SubmitTaskDeliverable(ctx, actor, task.ID, domain.DeliverableSubmission{RawOutput: "invalid but preserved output", Version: 2})
	if err != nil || second.ValidationStatus != "invalid" {
		t.Fatal(second, err)
	}
	empty, err := store.ListTaskRevisionRequests(ctx, actor, task.ID)
	if err != nil || empty.Version != 3 || empty.Items == nil || len(empty.Items) != 0 {
		t.Fatal("empty snapshot", empty, err)
	}
	firstNote, err := store.CreateTaskRevisionRequest(ctx, actor, task.ID, domain.RevisionRequestSubmission{DeliverableID: first.ID, Body: " \tClarify the historical evidence.\n", Version: 3})
	if err != nil || firstNote.ID == "" || firstNote.TaskID != task.ID || firstNote.DeliverableID != first.ID || firstNote.DeliverableRevision != first.Revision || firstNote.AuthorID != actor.UserID || firstNote.Body != "Clarify the historical evidence." || firstNote.CreatedAt.IsZero() {
		t.Fatal("historical version binding", firstNote, err)
	}
	secondNote, err := store.CreateTaskRevisionRequest(ctx, actor, task.ID, domain.RevisionRequestSubmission{DeliverableID: second.ID, Body: "Please repair this invalid output.", Version: 4})
	if err != nil || secondNote.DeliverableRevision != second.Revision || secondNote.DeliverableID != second.ID {
		t.Fatal("invalid output cannot receive feedback", secondNote, err)
	}
	current, err := store.GetTask(ctx, actor, task.ID)
	if err != nil || current.Version != 5 || current.Status != task.Status || current.AssignedAgentID != task.AssignedAgentID {
		t.Fatal("feedback changed lifecycle or assignment", current, err)
	}
	for _, original := range []domain.DeliverableVersion{first, second} {
		got, err := store.GetTaskDeliverable(ctx, actor, task.ID, original.ID)
		if err != nil || !reflect.DeepEqual(got, original) {
			t.Fatal("output changed", got, original, err)
		}
	}
	list, err := store.ListTaskRevisionRequests(ctx, actor, task.ID)
	if err != nil || list.Version != 5 || len(list.Items) != 2 || list.Items[0].ID != secondNote.ID || list.Items[1].ID != firstNote.ID || list.Items[1].Body != firstNote.Body {
		t.Fatal("snapshot", list, err)
	}
	events, err := store.TaskEvents(ctx, actor, task.ID)
	if err != nil || len(events) != 5 {
		t.Fatal("event count", events, err)
	}
	for i, note := range []domain.TaskRevisionRequest{firstNote, secondNote} {
		event := events[3+i]
		if event.Event != "revision_requested" || event.Version != int64(4+i) || event.Reason != note.ID || event.ActorID != actor.UserID || event.FromStatus != task.Status || event.ToStatus != task.Status || strings.Contains(event.Reason, note.Body) {
			t.Fatal("event must contain note ID, not its body", event)
		}
	}
	var audits, authoredAudits int
	if err = store.Pool.QueryRow(ctx, `SELECT count(*),count(*) FILTER (WHERE user_id=$3) FROM audit_logs WHERE org_id=$1 AND resource_id=$2 AND event='task.revision_requested'`, actor.OrgID, task.ID, actor.UserID).Scan(&audits, &authoredAudits); err != nil || audits != 2 || authoredAudits != 2 {
		t.Fatal("exactly one audit per request", audits, authoredAudits, err)
	}

	// Every failed direct repository call must leave all durable task state alone.
	snapshot := func() string {
		t.Helper()
		var out string
		err := store.Pool.QueryRow(ctx, `SELECT jsonb_build_object(
 'task',to_jsonb(t),
 'notes',(SELECT count(*) FROM task_revision_requests n WHERE n.org_id=t.org_id AND n.task_id=t.id),
 'outputs',(SELECT count(*) FROM task_deliverable_versions v WHERE v.org_id=t.org_id AND v.task_id=t.id),
 'events',(SELECT count(*) FROM task_events e WHERE e.org_id=t.org_id AND e.task_id=t.id),
 'audits',(SELECT count(*) FROM audit_logs a WHERE a.org_id=t.org_id AND a.resource_id=t.id))::text FROM tasks t WHERE t.org_id=$1 AND t.id=$2`, actor.OrgID, task.ID).Scan(&out)
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	baseline := snapshot()
	assertRejected := func(a domain.Actor, parent string, in domain.RevisionRequestSubmission, want error) {
		t.Helper()
		if _, err := store.CreateTaskRevisionRequest(ctx, a, parent, in); !errors.Is(err, want) {
			t.Fatalf("want %v, got %v", want, err)
		}
		if got := snapshot(); got != baseline {
			t.Fatalf("failed write changed state: %s != %s", got, baseline)
		}
	}
	input := domain.RevisionRequestSubmission{DeliverableID: first.ID, Body: "Another review note", Version: 5}
	for _, parent := range []string{uuid.NewString(), "not-a-uuid"} {
		assertRejected(actor, parent, input, domain.ErrNotFound)
		if _, err = store.ListTaskRevisionRequests(ctx, actor, parent); !errors.Is(err, domain.ErrNotFound) {
			t.Fatal("missing or malformed parent snapshot", err)
		}
	}
	for _, version := range []int64{0, 3, 4, 6} {
		stale := input
		stale.Version = version
		assertRejected(actor, task.ID, stale, domain.ErrTaskConflict)
	}
	for _, body := range []string{" \t\n", strings.Repeat("界", 4001), "bad\x00note", string([]byte{0xff})} {
		invalid := input
		invalid.Body = body
		assertRejected(actor, task.ID, invalid, domain.ErrInvalid)
	}
	for _, target := range []string{uuid.NewString(), "not-a-uuid"} {
		missing := input
		missing.DeliverableID = target
		assertRejected(actor, task.ID, missing, domain.ErrNotFound)
	}
	otherTask, err := service.Create(ctx, actor, domain.Task{BrandID: brand.ID, Title: "Wrong parent"})
	if err != nil {
		t.Fatal(err)
	}
	otherOutput, err := store.SubmitTaskDeliverable(ctx, actor, otherTask.ID, domain.DeliverableSubmission{RawOutput: "other output", Version: 1})
	if err != nil {
		t.Fatal(err)
	}
	wrongParent := input
	wrongParent.DeliverableID = otherOutput.ID
	assertRejected(actor, task.ID, wrongParent, domain.ErrNotFound)
	if _, err = store.ListTaskRevisionRequests(ctx, actor, uuid.NewString()); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("missing parent snapshot", err)
	}
	other, err := store.Bootstrap(ctx, "https://issuer.test/", uuid.NewString(), "Other tenant")
	if err != nil {
		t.Fatal(err)
	}
	otherBrand, err := (application.Brands{Repo: store}).Create(ctx, other, domain.Brand{Name: "Other tenant brand"})
	if err != nil {
		t.Fatal(err)
	}
	tenantTask, err := service.Create(ctx, other, domain.Task{BrandID: otherBrand.ID, Title: "Other tenant output"})
	if err != nil {
		t.Fatal(err)
	}
	tenantOutput, err := store.SubmitTaskDeliverable(ctx, other, tenantTask.ID, domain.DeliverableSubmission{RawOutput: "tenant output", Version: 1})
	if err != nil {
		t.Fatal(err)
	}
	crossTenant := input
	crossTenant.DeliverableID = tenantOutput.ID
	assertRejected(actor, task.ID, crossTenant, domain.ErrNotFound)
	assertRejected(other, task.ID, input, domain.ErrNotFound)
	if _, err = store.ListTaskRevisionRequests(ctx, other, task.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("cross-tenant list", err)
	}

	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := store.Pool.Exec(ctx, query, args...); err != nil {
			t.Fatal(err)
		}
	}
	// The repository must use the current database role, not actor.Role.
	for _, role := range []string{"viewer", "approver", "analyst"} {
		exec(`UPDATE user_roles SET role_id=$3 WHERE org_id=$1 AND user_id=$2`, actor.OrgID, actor.UserID, role)
		assertRejected(actor, task.ID, input, domain.ErrForbidden)
		if _, err = store.ListTaskRevisionRequests(ctx, actor, task.ID); err != nil {
			t.Fatal("read role cannot read", role, err)
		}
	}
	exec(`UPDATE user_roles SET role_id='owner' WHERE org_id=$1 AND user_id=$2`, actor.OrgID, actor.UserID)
	for _, query := range []string{
		`UPDATE user_brand_access SET active=false WHERE org_id=$1 AND user_id=$2`,
		`UPDATE organization_members SET active=false WHERE org_id=$1 AND user_id=$2`,
	} {
		exec(query, actor.OrgID, actor.UserID)
		assertRejected(actor, task.ID, input, domain.ErrNotFound)
		if _, err = store.ListTaskRevisionRequests(ctx, actor, task.ID); !errors.Is(err, domain.ErrNotFound) {
			t.Fatal("revoked read access", err)
		}
		exec(strings.Replace(query, "active=false", "active=true", 1), actor.OrgID, actor.UserID)
	}
	for _, status := range []domain.TaskStatus{domain.TaskQueued, domain.TaskRunning, domain.TaskWaitingInput, domain.TaskWaitingApproval, domain.TaskCompleted, domain.TaskFailed, domain.TaskCancelled} {
		exec(`UPDATE tasks SET status=$3 WHERE org_id=$1 AND id=$2`, actor.OrgID, task.ID, status)
		baseline = snapshot()
		assertRejected(actor, task.ID, input, domain.ErrInvalid)
	}
	exec(`UPDATE tasks SET status=$3 WHERE org_id=$1 AND id=$2`, actor.OrgID, task.ID, task.Status)
	baseline = snapshot()
	// Database-level bindings independently reject cross-parent/tenant and forged revisions.
	for _, target := range []struct {
		id       string
		revision int64
	}{{otherOutput.ID, otherOutput.Revision}, {tenantOutput.ID, tenantOutput.Revision}, {first.ID, first.Revision + 99}} {
		if _, err = store.Pool.Exec(ctx, `INSERT INTO task_revision_requests(org_id,id,task_id,deliverable_id,deliverable_revision,author_id,body) VALUES($1,$2,$3,$4,$5,$6,'forged binding')`, actor.OrgID, uuid.NewString(), task.ID, target.id, target.revision, actor.UserID); err == nil {
			t.Fatal("database accepted invalid target binding", target)
		}
	}
	if _, err = store.Pool.Exec(ctx, `INSERT INTO task_revision_requests(org_id,id,task_id,deliverable_id,deliverable_revision,author_id,body) VALUES($1,$2,$3,$4,$5,$6,'forged author')`, actor.OrgID, uuid.NewString(), task.ID, first.ID, first.Revision, other.UserID); err == nil {
		t.Fatal("database accepted non-member author")
	}
	for _, query := range []string{
		`UPDATE task_revision_requests SET body='overwritten' WHERE org_id=$1 AND id=$2`,
		`DELETE FROM task_revision_requests WHERE org_id=$1 AND id=$2`,
	} {
		if _, err = store.Pool.Exec(ctx, query, actor.OrgID, firstNote.ID); err == nil {
			t.Fatal("immutable request mutation succeeded", query)
		}
	}
	if snapshot() != baseline {
		t.Fatal("rejected SQL changed task state")
	}
	unchanged, err := store.ListTaskRevisionRequests(ctx, actor, task.ID)
	if err != nil || !reflect.DeepEqual(unchanged, list) {
		t.Fatal("failed writes changed notes", unchanged, err)
	}
}

func TestRealRevisionRequestListLatestHundredHasDeterministicTies(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()
	actor, err := store.Bootstrap(ctx, "https://issuer.test/", uuid.NewString(), "Revision list owner")
	if err != nil {
		t.Fatal(err)
	}
	brand, err := (application.Brands{Repo: store}).Create(ctx, actor, domain.Brand{Name: "Revision list fixture"})
	if err != nil {
		t.Fatal(err)
	}
	task, err := (application.Tasks{Repo: store}).Create(ctx, actor, domain.Task{BrandID: brand.ID, Title: "Revision list"})
	if err != nil {
		t.Fatal(err)
	}
	output, err := store.SubmitTaskDeliverable(ctx, actor, task.ID, domain.DeliverableSubmission{RawOutput: "preserve", Version: 1})
	if err != nil {
		t.Fatal(err)
	}
	ids := make([]string, 105)
	for i := range ids {
		ids[i] = uuid.NewString()
		_, err = store.Pool.Exec(ctx, `INSERT INTO task_revision_requests(org_id,id,task_id,deliverable_id,deliverable_revision,author_id,body,created_at) VALUES($1,$2,$3,$4,$5,$6,'tie note','2026-01-01T00:00:00Z')`, actor.OrgID, ids[i], task.ID, output.ID, output.Revision, actor.UserID)
		if err != nil {
			t.Fatal(err)
		}
	}
	latestID := uuid.NewString()
	if _, err = store.Pool.Exec(ctx, `INSERT INTO task_revision_requests(org_id,id,task_id,deliverable_id,deliverable_revision,author_id,body,created_at) VALUES($1,$2,$3,$4,$5,$6,'latest note','2026-01-02T00:00:00Z')`, actor.OrgID, latestID, task.ID, output.ID, output.Revision, actor.UserID); err != nil {
		t.Fatal(err)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(ids)))
	list, err := store.ListTaskRevisionRequests(ctx, actor, task.ID)
	if err != nil || list.Version != 2 || len(list.Items) != 100 || list.Items[0].ID != latestID {
		t.Fatal(list, err)
	}
	for i, note := range list.Items[1:] {
		if note.ID != ids[i] || note.Body != "tie note" || note.TaskID != task.ID || note.DeliverableID != output.ID || note.DeliverableRevision != output.Revision || note.AuthorID != actor.UserID || note.CreatedAt.IsZero() {
			t.Fatal("incorrect tie order or metadata", i, note)
		}
	}
}
