package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/menerjangbadai/dashboard/internal/application"
	"github.com/menerjangbadai/dashboard/internal/domain"
)

// Deferred database suite: testStore requires runner-provided disposable DSN.
func TestRealTaskPersistenceIsolationAndVersion(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()
	actor, err := store.Bootstrap(ctx, "https://issuer.test/", uuid.NewString(), "Task owner")
	if err != nil {
		t.Fatal(err)
	}
	brand, err := (application.Brands{Repo: store}).Create(ctx, actor, domain.Brand{Name: "Task fixture"})
	if err != nil {
		t.Fatal(err)
	}
	service := application.Tasks{Repo: store}
	task, err := service.Create(ctx, actor, domain.Task{BrandID: brand.ID, Title: "Brief"})
	if err != nil || task.Version != 1 {
		t.Fatalf("create: %+v %v", task, err)
	}
	task.Title = "Updated brief"
	stale := task
	task, err = service.Update(ctx, actor, task)
	if err != nil || task.Version != 2 {
		t.Fatalf("update: %+v %v", task, err)
	}
	if _, err = service.Update(ctx, actor, stale); !errors.Is(err, domain.ErrTaskConflict) {
		t.Fatal("stale edit accepted", err)
	}
	task, err = service.Transition(ctx, actor, task.ID, domain.TaskTransition{Status: domain.TaskReady, Version: task.Version})
	if err != nil {
		t.Fatal(err)
	}
	task.AssignedAgentID = domain.Agents()[4].ID
	task, err = service.Update(ctx, actor, task)
	if err != nil || task.Status != domain.TaskAssigned {
		t.Fatalf("assignment: %+v %v", task, err)
	}
	if _, err = service.Transition(ctx, actor, task.ID, domain.TaskTransition{Status: domain.TaskRunning, Version: task.Version}); !errors.Is(err, domain.ErrTaskExecutionDisabled) {
		t.Fatal("fake execution", err)
	}
	events, err := service.Events(ctx, actor, task.ID)
	if err != nil || len(events) != 4 {
		t.Fatalf("history count %d: %v", len(events), err)
	}
	other, err := store.Bootstrap(ctx, "https://issuer.test/", uuid.NewString(), "Other owner")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.Get(ctx, other, task.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("cross-org read", err)
	}
	if _, err = service.Events(ctx, other, task.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("cross-org events", err)
	}
}
