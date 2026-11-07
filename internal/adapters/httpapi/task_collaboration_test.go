package httpapi

import (
	"context"
	"github.com/menerjangbadai/dashboard/internal/application"
	"github.com/menerjangbadai/dashboard/internal/domain"
	"net/http"
	"strings"
	"testing"
	"time"
)

type collaborationHTTPRepo struct {
	taskHTTPRepo
	comments []domain.TaskComment
}

func (s *collaborationHTTPRepo) AddTaskComment(_ context.Context, _ domain.Actor, id string, in domain.TaskComment) (domain.TaskComment, error) {
	in.ID = "comment-fixture"
	in.TaskID = id
	in.AuthorID = "user"
	in.CreatedAt = time.Now()
	s.comments = append(s.comments, in)
	return in, nil
}
func (s *collaborationHTTPRepo) ListTaskComments(context.Context, domain.Actor, string) ([]domain.TaskComment, error) {
	return s.comments, nil
}
func TestTaskHTTPDeadlineAndComments(t *testing.T) {
	repo := &collaborationHTTPRepo{comments: []domain.TaskComment{}}
	h := New(Config{PublicURL: "https://app.test", CommitSHA: strings.Repeat("a", 40)}, application.Brands{}, &authMemory{}, provider{}, func(context.Context) error { return nil }, NewTaskHandler(application.Tasks{Repo: repo}))
	csrf := application.CSRF("opaque-test-session")
	w := call(h, "POST", "/api/tasks", `{"brand_id":"d0e0ddbb-23f8-4fe7-9be7-e53dd92e9860","title":"Brief","deadline_at":"2027-01-02T09:30:00+07:00"}`, "https://app.test", csrf, true)
	if w.Code != 201 || repo.task.DeadlineAt == nil {
		t.Fatalf("deadline create: %d %s", w.Code, w.Body.String())
	}
	path := "/api/tasks/" + repo.task.ID
	w = call(h, "PATCH", path, `{"version":1,"title":"Clarification brief"}`, "https://app.test", csrf, true)
	if w.Code != 200 || repo.task.DeadlineAt == nil || repo.task.DeadlineAt.Format(time.RFC3339) != "2027-01-02T09:30:00+07:00" {
		t.Fatalf("omitted deadline changed: %d %s", w.Code, w.Body.String())
	}
	w = call(h, "PATCH", path, `{"version":2,"deadline_at":null}`, "https://app.test", csrf, true)
	if w.Code != 200 || repo.task.DeadlineAt != nil {
		t.Fatalf("deadline clear: %d %s", w.Code, w.Body.String())
	}
	w = call(h, "POST", path+"/comments", `{"body":"  Who is our audience?  ","kind":"clarification"}`, "https://app.test", csrf, true)
	if w.Code != 201 || len(repo.comments) != 1 || repo.comments[0].Body != "Who is our audience?" {
		t.Fatalf("comment create: %d %s", w.Code, w.Body.String())
	}
	if repo.task.Status != domain.TaskDraft || repo.task.Version != 3 {
		t.Fatal("comment mutated task lifecycle")
	}
	w = call(h, "POST", path+"/comments", `{"body":"not an approval","kind":"approval"}`, "https://app.test", csrf, true)
	if w.Code != 422 || len(repo.comments) != 1 {
		t.Fatal("comment accepted approval", w.Code)
	}
	w = call(h, "GET", path+"/comments", "", "", "", true)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "Who is our audience?") {
		t.Fatalf("comments: %d %s", w.Code, w.Body.String())
	}
	w = call(h, "POST", path+"/comments", `{"body":"forged","author_id":"someone"}`, "https://app.test", csrf, true)
	if w.Code != 400 || len(repo.comments) != 1 {
		t.Fatal("forged author allowed", w.Code)
	}
	w = call(h, "POST", path+"/comments", `{"body":"bad csrf"}`, "https://evil.test", csrf, true)
	if w.Code != 403 || len(repo.comments) != 1 {
		t.Fatal("csrf bypass", w.Code)
	}
	w = call(h, "GET", path+"/comments", "", "", "", false)
	if w.Code != http.StatusUnauthorized {
		t.Fatal("unauthenticated comments", w.Code)
	}
}
