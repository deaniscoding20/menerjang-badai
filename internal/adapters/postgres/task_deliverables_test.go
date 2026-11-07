package postgres

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/menerjangbadai/dashboard/internal/application"
	"github.com/menerjangbadai/dashboard/internal/domain"
	"testing"
)

var _ application.TaskDeliverableRepository = (*Store)(nil)

// Deferred persistence acceptance: no shared/production database fallback.
func TestRealVersionedDeliverablesPreserveRawAndRejectStale(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()
	actor, err := store.Bootstrap(ctx, "https://issuer.test/", uuid.NewString(), "Deliverable owner")
	if err != nil {
		t.Fatal(err)
	}
	brand, err := (application.Brands{Repo: store}).Create(ctx, actor, domain.Brand{Name: "Deliverable fixture"})
	if err != nil {
		t.Fatal(err)
	}
	service := application.Tasks{Repo: store}
	task, err := service.Create(ctx, actor, domain.Task{BrandID: brand.ID, Title: "Research"})
	if err != nil {
		t.Fatal(err)
	}
	valid := ` {"schema_version":1,"kind":"research_report","title":"Audience","sections":[{"heading":"Finding","body":"Needs review"}]} `
	first, err := service.SubmitDeliverable(ctx, actor, task.ID, domain.DeliverableSubmission{RawOutput: valid, Version: task.Version})
	if err != nil || first.Revision != 1 || first.RawOutput != valid || first.Structured == nil || first.ValidationStatus != "validated" || first.AuthorID != actor.UserID || first.Origin != "human_import" {
		t.Fatal(first, err)
	}
	if _, err = service.SubmitDeliverable(ctx, actor, task.ID, domain.DeliverableSubmission{RawOutput: "stale", Version: 1}); !errors.Is(err, domain.ErrTaskConflict) {
		t.Fatal("stale accepted", err)
	}
	task, err = service.Get(ctx, actor, task.ID)
	if err != nil || task.Version != 2 || task.Status != domain.TaskDraft {
		t.Fatal("import changed lifecycle", task, err)
	}
	second, err := service.SubmitDeliverable(ctx, actor, task.ID, domain.DeliverableSubmission{RawOutput: "malformed but preserved", Version: task.Version})
	if err != nil || second.Revision != 2 || second.Structured != nil || second.ValidationStatus != "invalid" {
		t.Fatal(second, err)
	}
	got, err := service.Deliverable(ctx, actor, task.ID, first.ID)
	if err != nil || got.RawOutput != valid || got.Revision != 1 {
		t.Fatal("first version overwritten", got, err)
	}
	list, err := service.Deliverables(ctx, actor, task.ID)
	if err != nil || list.Version != 3 || len(list.Items) != 2 || list.Items[0].Revision != 2 || list.Items[0].Kind != "" {
		t.Fatal(list, err)
	}
	events, err := service.Events(ctx, actor, task.ID)
	if err != nil || len(events) != 3 || events[2].Event != "deliverable_submitted" {
		t.Fatal(events, err)
	}
	var count int
	if err = store.Pool.QueryRow(ctx, `SELECT count(*) FROM audit_logs WHERE org_id=$1 AND resource_id=$2 AND event='task.deliverable_submitted'`, actor.OrgID, task.ID).Scan(&count); err != nil || count != 2 {
		t.Fatal("audit", count, err)
	}
	// Binding both IDs prevents reading another task's version through this task route.
	otherTask, err := service.Create(ctx, actor, domain.Task{BrandID: brand.ID, Title: "Other task"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.Deliverable(ctx, actor, otherTask.ID, first.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("wrong parent", err)
	}
	otherActor, err := store.Bootstrap(ctx, "https://issuer.test/", uuid.NewString(), "Other owner")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.Deliverable(ctx, otherActor, task.ID, first.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("cross org", err)
	}
	if _, err = store.Pool.Exec(ctx, `UPDATE task_deliverable_versions SET raw_output='overwrite' WHERE org_id=$1 AND id=$2`, actor.OrgID, first.ID); err == nil {
		t.Fatal("immutable version update succeeded")
	}
}
