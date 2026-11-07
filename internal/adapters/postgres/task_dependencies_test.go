package postgres

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/menerjangbadai/dashboard/internal/application"
	"github.com/menerjangbadai/dashboard/internal/domain"
	"sync"
	"testing"
)

var _ application.TaskDependencyRepository = (*Store)(nil)

// Deferred: requires testStore's explicitly enabled disposable PostgreSQL.
func TestRealTaskDependencyPersistenceCycleAndAudit(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()
	actor, err := store.Bootstrap(ctx, "https://issuer.test/", uuid.NewString(), "Dependency owner")
	if err != nil {
		t.Fatal(err)
	}
	brand, err := (application.Brands{Repo: store}).Create(ctx, actor, domain.Brand{Name: "Dependency fixture"})
	if err != nil {
		t.Fatal(err)
	}
	service := application.Tasks{Repo: store}
	makeTask := func(title string) domain.Task {
		t.Helper()
		task, e := service.Create(ctx, actor, domain.Task{BrandID: brand.ID, Title: title})
		if e != nil {
			t.Fatal(e)
		}
		return task
	}
	a, b, c := makeTask("A"), makeTask("B"), makeTask("C")
	change := func(task domain.Task, target domain.Task, add bool) (domain.Task, error) {
		return service.ChangeDependency(ctx, actor, task.ID, domain.TaskDependencyChange{DependencyID: target.ID, Version: task.Version}, add)
	}
	stale := a
	a, err = change(a, b, true)
	if err != nil || a.Version != 2 || a.Status != domain.TaskDraft {
		t.Fatal("add", a, err)
	}
	snapshot, err := service.Dependencies(ctx, actor, a.ID)
	if err != nil || snapshot.Version != a.Version || len(snapshot.Items) != 1 || snapshot.Items[0].TaskID != b.ID {
		t.Fatal(snapshot, err)
	}
	same, err := change(a, b, true)
	if err != nil || same.Version != a.Version {
		t.Fatal("duplicate should be no-op", same, err)
	}
	if _, err = change(stale, c, true); !errors.Is(err, domain.ErrTaskConflict) {
		t.Fatal("stale", err)
	}
	if _, err = change(stale, b, true); !errors.Is(err, domain.ErrTaskConflict) {
		t.Fatal("stale duplicate add", err)
	}
	if _, err = change(stale, c, false); !errors.Is(err, domain.ErrTaskConflict) {
		t.Fatal("stale missing remove", err)
	}
	same, err = change(a, c, false)
	if err != nil || same.Version != a.Version || same.Status != a.Status {
		t.Fatal("missing remove should be no-op", same, err)
	}
	noopEvents, err := service.Events(ctx, actor, a.ID)
	if err != nil || len(noopEvents) != 2 {
		t.Fatal("no-op changed history", noopEvents, err)
	}
	var noopAudits int
	if err = store.Pool.QueryRow(ctx, `SELECT count(*) FROM audit_logs WHERE org_id=$1 AND resource_id=$2 AND event IN ('task.dependency_added','task.dependency_removed')`, actor.OrgID, a.ID).Scan(&noopAudits); err != nil || noopAudits != 1 {
		t.Fatal("no-op changed audit", noopAudits, err)
	}
	b, err = change(b, c, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = change(c, a, true); !errors.Is(err, domain.ErrDependencyCycle) {
		t.Fatal("transitive cycle", err)
	}
	snapshot, err = service.Dependencies(ctx, actor, c.ID)
	if err != nil || snapshot.Version != 1 || len(snapshot.Items) != 0 {
		t.Fatal("failed mutation leaked", snapshot, err)
	}
	a, err = change(a, b, false)
	if err != nil || a.Version != 3 || a.Status != domain.TaskDraft {
		t.Fatal("remove", a, err)
	}
	events, err := service.Events(ctx, actor, a.ID)
	if err != nil || len(events) != 3 || events[1].Event != "dependency_added" || events[2].Event != "dependency_removed" {
		t.Fatal(events, err)
	}
	var auditCount int
	if err = store.Pool.QueryRow(ctx, `SELECT count(*) FROM audit_logs WHERE org_id=$1 AND resource_id=$2 AND event IN ('task.dependency_added','task.dependency_removed')`, actor.OrgID, a.ID).Scan(&auditCount); err != nil || auditCount != 2 {
		t.Fatal("audit", auditCount, err)
	}
	other, err := (application.Brands{Repo: store}).Create(ctx, actor, domain.Brand{Name: "Other dependency brand"})
	if err != nil {
		t.Fatal(err)
	}
	target, err := service.Create(ctx, actor, domain.Task{BrandID: other.ID, Title: "Other brand"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = change(a, target, true); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("cross brand", err)
	}
	// Opposite simultaneous writes to different source rows must not both commit.
	x, y := makeTask("X"), makeTask("Y")
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for _, pair := range [][2]domain.Task{{x, y}, {y, x}} {
		wg.Add(1)
		go func(pair [2]domain.Task) {
			defer wg.Done()
			<-start
			_, e := change(pair[0], pair[1], true)
			results <- e
		}(pair)
	}
	close(start)
	wg.Wait()
	close(results)
	successes, cycles := 0, 0
	for e := range results {
		if e == nil {
			successes++
		} else if errors.Is(e, domain.ErrDependencyCycle) {
			cycles++
		} else {
			t.Fatal("concurrent graph", e)
		}
	}
	if successes != 1 || cycles != 1 {
		t.Fatal("concurrent cycle accepted", successes, cycles)
	}
}
