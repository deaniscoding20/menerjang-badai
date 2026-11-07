package application

import (
	"context"
	"errors"
	"github.com/menerjangbadai/dashboard/internal/domain"
	"testing"
)

type commentRepo struct {
	TaskRepository
	role       string
	taskDenied bool
	wrote      bool
	comment    domain.TaskComment
}

func (r *commentRepo) Membership(_ context.Context, a domain.Actor) (domain.Actor, error) {
	a.Role = r.role
	return a, nil
}
func (r *commentRepo) GetTask(context.Context, domain.Actor, string) (domain.Task, error) {
	if r.taskDenied {
		return domain.Task{}, domain.ErrNotFound
	}
	return domain.Task{}, nil
}
func (r *commentRepo) AddTaskComment(_ context.Context, _ domain.Actor, id string, in domain.TaskComment) (domain.TaskComment, error) {
	r.wrote = true
	in.TaskID = id
	r.comment = in
	return in, nil
}
func (r *commentRepo) ListTaskComments(context.Context, domain.Actor, string) ([]domain.TaskComment, error) {
	return []domain.TaskComment{}, nil
}
func TestCommentRequiresCurrentWriteRoleAndVisibleTask(t *testing.T) {
	repo := &commentRepo{role: "viewer"}
	service := Tasks{Repo: repo}
	id := "dbf244dd-5907-4308-bdb4-6fe2d3d58cd8"
	if _, err := service.AddComment(context.Background(), domain.Actor{Role: "owner"}, id, domain.TaskComment{Body: "Note"}); !errors.Is(err, domain.ErrForbidden) || repo.wrote {
		t.Fatal("read-only write", err)
	}
	repo.role = "editor"
	repo.taskDenied = true
	if _, err := service.AddComment(context.Background(), domain.Actor{}, id, domain.TaskComment{Body: "Note"}); !errors.Is(err, domain.ErrNotFound) || repo.wrote {
		t.Fatal("hidden task write", err)
	}
	repo.taskDenied = false
	comment, err := service.AddComment(context.Background(), domain.Actor{}, id, domain.TaskComment{Body: "  Need target audience  ", Kind: "clarification"})
	if err != nil || !repo.wrote || comment.Body != "Need target audience" || comment.Kind != "clarification" {
		t.Fatalf("valid note %+v %v", comment, err)
	}
}
