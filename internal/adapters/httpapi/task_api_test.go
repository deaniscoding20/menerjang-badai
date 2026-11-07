package httpapi

import (
	"context"
	"encoding/json"
	"github.com/menerjangbadai/dashboard/internal/application"
	"github.com/menerjangbadai/dashboard/internal/domain"
	"net/http"
	"strings"
	"testing"
	"time"
)

type taskHTTPRepo struct {
	application.TaskRepository
	task domain.Task
}

func (s *taskHTTPRepo) Membership(_ context.Context, a domain.Actor) (domain.Actor, error) {
	a.Role = "owner"
	return a, nil
}
func (s *taskHTTPRepo) GetBrand(context.Context, domain.Actor, string) (domain.Brand, error) {
	return domain.Brand{Status: "active"}, nil
}
func (s *taskHTTPRepo) CreateTask(_ context.Context, _ domain.Actor, in domain.Task) (domain.Task, error) {
	in.ID = "dbf244dd-5907-4308-bdb4-6fe2d3d58cd8"
	in.Version = 1
	in.CreatedAt = time.Now()
	in.UpdatedAt = in.CreatedAt
	s.task = in
	return in, nil
}
func (s *taskHTTPRepo) GetTask(context.Context, domain.Actor, string) (domain.Task, error) {
	return s.task, nil
}
func (s *taskHTTPRepo) UpdateTask(_ context.Context, _ domain.Actor, in domain.Task) (domain.Task, error) {
	if in.Version != s.task.Version {
		return domain.Task{}, domain.ErrTaskConflict
	}
	in.Version++
	s.task = in
	return in, nil
}
func (s *taskHTTPRepo) TransitionTask(_ context.Context, _ domain.Actor, _ string, in domain.TaskTransition) (domain.Task, error) {
	next, err := s.task.ApplyTransition(in)
	if err != nil {
		return domain.Task{}, err
	}
	next.Version++
	s.task = next
	return next, nil
}

func TestTaskHTTPCreateUpdateAndExecutionIsNotFaked(t *testing.T) {
	repo := &taskHTTPRepo{}
	handler := New(Config{PublicURL: "https://app.test", CommitSHA: strings.Repeat("a", 40)}, application.Brands{}, &authMemory{}, provider{}, func(context.Context) error { return nil }, NewTaskHandler(application.Tasks{Repo: repo}))
	csrf := application.CSRF("opaque-test-session")
	response := call(handler, "POST", "/api/tasks", `{"brand_id":"d0e0ddbb-23f8-4fe7-9be7-e53dd92e9860","title":"Brief","agent_id":null}`, "https://app.test", csrf, true)
	if response.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", response.Code, response.Body.String())
	}
	var task domain.Task
	if json.Unmarshal(response.Body.Bytes(), &task) != nil || task.Version != 1 || task.Status != domain.TaskDraft {
		t.Fatalf("draft response %s", response.Body.String())
	}
	response = call(handler, "PATCH", "/api/tasks/"+task.ID, `{"title":"Edited","version":1}`, "https://app.test", csrf, true)
	if response.Code != 200 || !strings.Contains(response.Body.String(), "Edited") {
		t.Fatalf("update: %d %s", response.Code, response.Body.String())
	}
	response = call(handler, "PATCH", "/api/tasks/"+task.ID, `{"title":"Stale","version":1}`, "https://app.test", csrf, true)
	if response.Code != 409 {
		t.Fatalf("stale edit: %d %s", response.Code, response.Body.String())
	}
	response = call(handler, "POST", "/api/tasks/"+task.ID+"/transition", `{"status":"running","version":2}`, "https://app.test", csrf, true)
	if response.Code != 409 || !strings.Contains(response.Body.String(), "execution_disabled") {
		t.Fatalf("unverified run: %d %s", response.Code, response.Body.String())
	}
	response = call(handler, "PATCH", "/api/tasks/"+task.ID, `{"title":"No version"}`, "https://app.test", csrf, true)
	if response.Code != 409 {
		t.Fatalf("missing version: %d %s", response.Code, response.Body.String())
	}
}
