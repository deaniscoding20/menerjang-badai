package postgres

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/menerjangbadai/dashboard/internal/application"
	"github.com/menerjangbadai/dashboard/internal/domain"
	"testing"
	"time"
)

// Deferred: only run using scripts/test_setup.py database on a disposable DB.
// This regression is not exercised by unit/fixture checks.
func TestRealReadyTaskMetadataPreservesAssignment(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()
	actor, err := store.Bootstrap(ctx, "https://issuer.test/", uuid.NewString(), "Ready metadata fixture")
	if err != nil {
		t.Fatal(err)
	}
	brand, err := (application.Brands{Repo: store}).Create(ctx, actor, domain.Brand{Name: "Ready metadata fixture"})
	if err != nil {
		t.Fatal(err)
	}
	service := application.Tasks{Repo: store}
	deadline := time.Date(2027, 1, 2, 9, 30, 0, 0, time.UTC)
	agentID := domain.Agents()[4].ID
	task, err := service.Create(ctx, actor, domain.Task{BrandID: brand.ID, Title: "Ready brief", AssignedAgentID: agentID, DeadlineAt: &deadline})
	if err != nil {
		t.Fatal(err)
	}
	task, err = service.Transition(ctx, actor, task.ID, domain.TaskTransition{Status: domain.TaskReady, Version: task.Version})
	if err != nil {
		t.Fatal(err)
	}
	for _, edit := range []string{"brief", "change deadline", "clear deadline"} {
		previous := task
		switch edit {
		case "brief":
			task.Title = "Updated ready brief"
			task.Objective = "Updated objective"
			task.Instructions = "Updated instructions"
		case "change deadline":
			later := deadline.Add(time.Hour)
			task.DeadlineAt = &later
		case "clear deadline":
			task.DeadlineAt = nil
		}
		wanted := task
		task, err = service.Update(ctx, actor, task)
		if err != nil || task.Status != domain.TaskReady || task.AssignedAgentID != agentID || task.Version != previous.Version+1 {
			t.Fatalf("%s changed assignment: %+v %v", edit, task, err)
		}
		persisted, err := service.Get(ctx, actor, task.ID)
		if err != nil || persisted.Status != domain.TaskReady || persisted.AssignedAgentID != agentID || persisted.Version != task.Version || persisted.Title != wanted.Title || persisted.Objective != wanted.Objective || persisted.Instructions != wanted.Instructions {
			t.Fatalf("%s not persisted: %+v %v", edit, persisted, err)
		}
		if (persisted.DeadlineAt == nil) != (wanted.DeadlineAt == nil) || (wanted.DeadlineAt != nil && !persisted.DeadlineAt.Equal(*wanted.DeadlineAt)) {
			t.Fatalf("%s changed deadline: got %v want %v", edit, persisted.DeadlineAt, wanted.DeadlineAt)
		}
		if _, err = service.Update(ctx, actor, previous); !errors.Is(err, domain.ErrTaskConflict) {
			t.Fatalf("%s stale update accepted: %v", edit, err)
		}
	}
	assigned, err := service.Transition(ctx, actor, task.ID, domain.TaskTransition{Status: domain.TaskAssigned, Version: task.Version})
	if err != nil || assigned.Status != domain.TaskAssigned || assigned.AssignedAgentID != agentID {
		t.Fatalf("explicit ready-to-assigned transition: %+v %v", assigned, err)
	}
}

// Deferred: only run using scripts/test_setup.py database on a disposable DB.
// Compilation is not evidence of PostgreSQL persistence or transaction correctness.
func TestRealTaskDeadlineAndCommentPersistence(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()
	actor, err := store.Bootstrap(ctx, "https://issuer.test/", uuid.NewString(), "Collaboration fixture")
	if err != nil {
		t.Fatal(err)
	}
	brand, err := (application.Brands{Repo: store}).Create(ctx, actor, domain.Brand{Name: "Collaboration fixture"})
	if err != nil {
		t.Fatal(err)
	}
	service := application.Tasks{Repo: store}
	deadline := time.Date(2027, 1, 2, 9, 30, 45, 123000000, time.FixedZone("UTC+7", 7*60*60))
	task, err := service.Create(ctx, actor, domain.Task{BrandID: brand.ID, Title: "Deadline fixture", DeadlineAt: &deadline})
	if err != nil || task.DeadlineAt == nil || !task.DeadlineAt.Equal(deadline) {
		t.Fatalf("create deadline: %+v %v", task, err)
	}
	// Unrelated edit retains the same instant; explicit removal still increments revision.
	task.Title = "Edited"
	task, err = service.Update(ctx, actor, task)
	if err != nil || task.DeadlineAt == nil || !task.DeadlineAt.Equal(deadline) {
		t.Fatalf("update deadline: %+v %v", task, err)
	}
	stale := task
	task.DeadlineAt = nil
	task, err = service.Update(ctx, actor, task)
	if err != nil || task.DeadlineAt != nil {
		t.Fatalf("clear deadline: %+v %v", task, err)
	}
	if _, err = service.Update(ctx, actor, stale); !errors.Is(err, domain.ErrTaskConflict) {
		t.Fatal("stale deadline accepted", err)
	}
	comment, err := service.AddComment(ctx, actor, task.ID, domain.TaskComment{Body: "  Target audience?  ", Kind: "clarification", AuthorID: uuid.NewString()})
	if err != nil || comment.AuthorID != actor.UserID || comment.Body != "Target audience?" || comment.ID == "" {
		t.Fatalf("comment: %+v %v", comment, err)
	}
	comments, err := service.Comments(ctx, actor, task.ID)
	if err != nil || len(comments) != 1 || comments[0].ID != comment.ID {
		t.Fatalf("comment persistence: %+v %v", comments, err)
	}
	current, err := service.Get(ctx, actor, task.ID)
	if err != nil || current.Status != task.Status || current.Version != task.Version {
		t.Fatal("comment changed task lifecycle", err)
	}
	var auditCount int
	if err = store.Pool.QueryRow(ctx, `SELECT count(*) FROM audit_logs WHERE org_id=$1 AND user_id=$2 AND resource_id=$3 AND event='task.comment_added'`, actor.OrgID, actor.UserID, task.ID).Scan(&auditCount); err != nil || auditCount != 1 {
		t.Fatal("comment audit missing", err)
	}
	other, err := store.Bootstrap(ctx, "https://issuer.test/", uuid.NewString(), "Other fixture")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.Comments(ctx, other, task.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("cross-org comments", err)
	}
	if _, err = service.AddComment(ctx, other, task.ID, domain.TaskComment{Body: "Denied"}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("cross-org write", err)
	}
	if _, err = store.Pool.Exec(ctx, `UPDATE user_roles SET role_id='viewer' WHERE org_id=$1 AND user_id=$2`, actor.OrgID, actor.UserID); err != nil {
		t.Fatal(err)
	}
	// A stale owner claim cannot bypass a current read-only role, even at the adapter.
	if _, err = store.AddTaskComment(ctx, actor, task.ID, domain.TaskComment{Body: "Denied"}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal("revoked write authority", err)
	}
	if comments, err = service.Comments(ctx, actor, task.ID); err != nil || len(comments) != 1 {
		t.Fatal("viewer cannot read", err)
	}
	if _, err = store.Pool.Exec(ctx, `UPDATE user_brand_access SET active=false WHERE org_id=$1 AND user_id=$2 AND brand_id=$3`, actor.OrgID, actor.UserID, brand.ID); err != nil {
		t.Fatal(err)
	}
	if comments, err = store.ListTaskComments(ctx, actor, task.ID); err != nil || len(comments) != 0 {
		t.Fatal("revoked grant leaked notes", err)
	}
}
