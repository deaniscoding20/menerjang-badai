package postgres

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/menerjangbadai/dashboard/internal/application"
	"github.com/menerjangbadai/dashboard/internal/domain"
)

// Deferred disposable-DB acceptance. This intentionally injects a constraint
// failure after earlier writes, not just an invalid input rejected before BEGIN.
// Compile-only checks do not establish that the SQL or rollback works.
func TestRealRevisionRequestRollsBackLateEventAndAuditFailures(t *testing.T) {
	for _, stage := range []string{"event", "audit"} {
		t.Run(stage, func(t *testing.T) {
			store := testStore(t)
			ctx := context.Background()
			actor, err := store.Bootstrap(ctx, "https://issuer.test/", uuid.NewString(), "Revision rollback owner")
			if err != nil {
				t.Fatal(err)
			}
			brand, err := (application.Brands{Repo: store}).Create(ctx, actor, domain.Brand{Name: "Revision rollback fixture"})
			if err != nil {
				t.Fatal(err)
			}
			task, err := (application.Tasks{Repo: store}).Create(ctx, actor, domain.Task{BrandID: brand.ID, Title: "Rollback feedback"})
			if err != nil {
				t.Fatal(err)
			}
			output, err := store.SubmitTaskDeliverable(ctx, actor, task.ID, domain.DeliverableSubmission{RawOutput: "immutable output", Version: task.Version})
			if err != nil {
				t.Fatal(err)
			}
			beforeTask, err := store.GetTask(ctx, actor, task.ID)
			if err != nil {
				t.Fatal(err)
			}
			beforeEvents, err := store.TaskEvents(ctx, actor, task.ID)
			if err != nil {
				t.Fatal(err)
			}
			auditCount := func() int64 {
				t.Helper()
				var count int64
				if err := store.Pool.QueryRow(ctx, `SELECT count(*) FROM audit_logs WHERE org_id=$1 AND resource_id=$2`, actor.OrgID, task.ID).Scan(&count); err != nil {
					t.Fatal(err)
				}
				return count
			}
			beforeAudits := auditCount()
			table, resourceColumn, event := "task_events", "task_id", "revision_requested"
			if stage == "audit" {
				table, resourceColumn, event = "audit_logs", "resource_id", "task.revision_requested"
			}
			constraint := "revision_fault_" + strings.ReplaceAll(uuid.NewString(), "-", "")
			// PostgreSQL format quotes identifiers/literals; do not interpolate input text.
			var statement string
			if err = store.Pool.QueryRow(ctx, `SELECT format('ALTER TABLE %I ADD CONSTRAINT %I CHECK (NOT (org_id=%L::uuid AND %I=%L::uuid AND event=%L)) NOT VALID', $1::text,$2::text,$3::text,$4::text,$5::text,$6::text)`, table, constraint, actor.OrgID, resourceColumn, task.ID, event).Scan(&statement); err != nil {
				t.Fatal(err)
			}
			if _, err = store.Pool.Exec(ctx, statement); err != nil {
				t.Fatal(err)
			}
			removeFault := func() error {
				_, err := store.Pool.Exec(ctx, "ALTER TABLE "+pgx.Identifier{table}.Sanitize()+" DROP CONSTRAINT IF EXISTS "+pgx.Identifier{constraint}.Sanitize())
				return err
			}
			// A subtest cleanup runs before testStore closes its connection pool.
			t.Cleanup(func() {
				if err := removeFault(); err != nil {
					t.Errorf("fault cleanup failed: %v", err)
				}
			})
			in := domain.RevisionRequestSubmission{DeliverableID: output.ID, Body: "Please clarify evidence", Version: beforeTask.Version}
			if _, err = store.CreateTaskRevisionRequest(ctx, actor, task.ID, in); err == nil {
				t.Fatal("injected late failure unexpectedly committed")
			}
			afterTask, err := store.GetTask(ctx, actor, task.ID)
			if err != nil || !reflect.DeepEqual(afterTask, beforeTask) {
				t.Fatal("task/version changed after rollback", afterTask, err)
			}
			afterEvents, err := store.TaskEvents(ctx, actor, task.ID)
			if err != nil || !reflect.DeepEqual(afterEvents, beforeEvents) {
				t.Fatal("events changed after rollback", afterEvents, err)
			}
			notes, err := store.ListTaskRevisionRequests(ctx, actor, task.ID)
			if err != nil || notes.Version != beforeTask.Version || len(notes.Items) != 0 || notes.Items == nil {
				t.Fatal("note survived rollback", notes, err)
			}
			if got := auditCount(); got != beforeAudits {
				t.Fatal("audit survived rollback", got, beforeAudits)
			}
			unchanged, err := store.GetTaskDeliverable(ctx, actor, task.ID, output.ID)
			if err != nil || !reflect.DeepEqual(unchanged, output) {
				t.Fatal("target output changed", unchanged, err)
			}
			if err = removeFault(); err != nil {
				t.Fatal(err)
			}
			// Reuse the same optimistic version: failure must not consume it.
			saved, err := store.CreateTaskRevisionRequest(ctx, actor, task.ID, in)
			if err != nil || saved.DeliverableID != output.ID || saved.DeliverableRevision != output.Revision {
				t.Fatal("recovery", saved, err)
			}
			recovered, err := store.GetTask(ctx, actor, task.ID)
			if err != nil || recovered.Version != beforeTask.Version+1 || recovered.Status != beforeTask.Status {
				t.Fatal("recovery task", recovered, err)
			}
			notes, err = store.ListTaskRevisionRequests(ctx, actor, task.ID)
			if err != nil || len(notes.Items) != 1 || notes.Items[0].ID != saved.ID || notes.Version != recovered.Version {
				t.Fatal("recovery note", notes, err)
			}
			events, err := store.TaskEvents(ctx, actor, task.ID)
			if err != nil || len(events) != len(beforeEvents)+1 || events[len(events)-1].Event != "revision_requested" || events[len(events)-1].Reason != saved.ID {
				t.Fatal("recovery event", events, err)
			}
			if got := auditCount(); got != beforeAudits+1 {
				t.Fatal("recovery audit", got, beforeAudits+1)
			}
		})
	}
}
