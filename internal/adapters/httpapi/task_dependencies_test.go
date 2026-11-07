package httpapi

import (
	"context"
	"github.com/menerjangbadai/dashboard/internal/application"
	"github.com/menerjangbadai/dashboard/internal/domain"
	"strings"
	"testing"
)

const httpDependencySource = "dbf244dd-5907-4308-bdb4-6fe2d3d58cd8"
const httpDependencyTarget = "89b899e1-5922-4053-9417-7eaf90d170f7"

type dependencyHTTPRepo struct {
	taskHTTPRepo
	cycle   bool
	changed bool
	role    string
}

func (r *dependencyHTTPRepo) Membership(_ context.Context, a domain.Actor) (domain.Actor, error) {
	a.Role = r.role
	return a, nil
}
func (r *dependencyHTTPRepo) GetTask(_ context.Context, _ domain.Actor, id string) (domain.Task, error) {
	task := r.task
	task.ID = id
	return task, nil
}
func (r *dependencyHTTPRepo) ListTaskDependencies(context.Context, domain.Actor, string) (domain.TaskDependencies, error) {
	return domain.TaskDependencies{Version: r.task.Version, Items: []domain.TaskDependency{{TaskID: httpDependencyTarget, Title: "Research", Status: domain.TaskDraft}}}, nil
}
func (r *dependencyHTTPRepo) ChangeTaskDependency(_ context.Context, _ domain.Actor, _ string, _ domain.TaskDependencyChange, _ bool) (domain.Task, error) {
	if r.cycle {
		return domain.Task{}, domain.ErrDependencyCycle
	}
	r.changed = true
	r.task.Version++
	return r.task, nil
}
func TestTaskDependencyHTTPContractSessionCSRFAndConflict(t *testing.T) {
	repo := &dependencyHTTPRepo{role: "editor", taskHTTPRepo: taskHTTPRepo{task: domain.Task{ID: httpDependencySource, BrandID: "d0e0ddbb-23f8-4fe7-9be7-e53dd92e9860", Status: domain.TaskReady, Version: 3}}}
	handler := New(Config{PublicURL: "https://app.test", CommitSHA: strings.Repeat("a", 40)}, application.Brands{}, &authMemory{}, provider{}, func(context.Context) error { return nil }, NewTaskHandler(application.Tasks{Repo: repo}))
	path := "/api/tasks/" + httpDependencySource + "/dependencies"
	csrf := application.CSRF("opaque-test-session")
	in := `{"dependency_id":"` + httpDependencyTarget + `","version":3}`
	if res := call(handler, "GET", path, "", "", "", false); res.Code != 401 {
		t.Fatal(res.Code, res.Body.String())
	}
	if res := call(handler, "GET", path, "", "", "", true); res.Code != 200 || !strings.Contains(res.Body.String(), `"version":3`) || !strings.Contains(res.Body.String(), `"title":"Research"`) {
		t.Fatal(res.Code, res.Body.String())
	}
	for _, method := range []string{"POST", "DELETE"} {
		if res := call(handler, method, path, in, "https://app.test", "bad", true); res.Code != 403 || repo.changed {
			t.Fatal("csrf", method, res.Code)
		}
	}
	if res := call(handler, "POST", path, `{"dependency_id":"`+httpDependencyTarget+`","version":3,"org_id":"forged"}`, "https://app.test", csrf, true); res.Code != 400 || repo.changed {
		t.Fatal("forged authority", res.Code)
	}
	if res := call(handler, "POST", path, `{"dependency_id":"`+httpDependencyTarget+`","version":2}`, "https://app.test", csrf, true); res.Code != 409 || repo.changed {
		t.Fatal("stale", res.Code)
	}
	repo.cycle = true
	if res := call(handler, "POST", path, in, "https://app.test", csrf, true); res.Code != 409 || !strings.Contains(res.Body.String(), "dependency_cycle") {
		t.Fatal("cycle", res.Code, res.Body.String())
	}
	repo.cycle = false
	repo.role = "viewer"
	if res := call(handler, "POST", path, in, "https://app.test", csrf, true); res.Code != 403 || repo.changed {
		t.Fatal("read-only", res.Code)
	}
	repo.role = "editor"
	if res := call(handler, "POST", path, in, "https://app.test", csrf, true); res.Code != 200 || !strings.Contains(res.Body.String(), `"version":4`) || !repo.changed {
		t.Fatal("add", res.Code, res.Body.String())
	}
	repo.changed = false
	if res := call(handler, "DELETE", path, `{"dependency_id":"`+httpDependencyTarget+`","version":4}`, "https://app.test", csrf, true); res.Code != 200 || !repo.changed || repo.task.Status != domain.TaskReady {
		t.Fatal("remove", res.Code, res.Body.String())
	}
}
