package httpapi

import (
	"context"
	"github.com/menerjangbadai/dashboard/internal/application"
	"github.com/menerjangbadai/dashboard/internal/domain"
	"net/http"
	"strings"
	"testing"
)

func TestTaskHTTPRejectsForgedIdentityFields(t *testing.T) {
	repo := &taskHTTPRepo{}
	h := New(Config{PublicURL: "https://app.test", CommitSHA: strings.Repeat("a", 40)}, application.Brands{}, &authMemory{}, provider{}, func(context.Context) error { return nil }, NewTaskHandler(application.Tasks{Repo: repo}))
	csrf := application.CSRF("opaque-test-session")
	response := call(h, "POST", "/api/tasks", `{"org_id":"forged","brand_id":"d0e0ddbb-23f8-4fe7-9be7-e53dd92e9860","title":"Brief"}`, "https://app.test", csrf, true)
	if response.Code != 400 || repo.task.ID != "" {
		t.Fatalf("forged field accepted: %d", response.Code)
	}
	response = call(h, "POST", "/api/tasks", `{"status":"completed","brand_id":"d0e0ddbb-23f8-4fe7-9be7-e53dd92e9860","title":"Brief"}`, "https://app.test", csrf, true)
	if response.Code != 400 || repo.task.ID != "" {
		t.Fatalf("forged status accepted: %d", response.Code)
	}
	// The handler cannot be mounted separately to bypass session middleware.
	raw := NewTaskHandler(application.Tasks{Repo: repo})
	response = call(raw, "GET", "/api/tasks", "", "", "", true)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("untrusted context accepted: %d", response.Code)
	}
}

func TestTaskDomainRejectsStaleAssignmentAndTerminalEdits(t *testing.T) {
	task, err := domain.NewTask(domain.Task{BrandID: "d0e0ddbb-23f8-4fe7-9be7-e53dd92e9860", Title: "Brief"})
	if err != nil {
		t.Fatal(err)
	}
	task.Version = 2
	if _, err = task.ApplyAssignment(domain.TaskAssignment{AgentID: domain.Agents()[0].ID, Version: 1}); err != domain.ErrTaskConflict {
		t.Fatal("stale assignment accepted", err)
	}
	task.Status = domain.TaskCancelled
	if _, err = task.ApplyAssignment(domain.TaskAssignment{AgentID: domain.Agents()[0].ID, Version: 2}); err != domain.ErrInvalid {
		t.Fatal("terminal edit accepted", err)
	}
}
