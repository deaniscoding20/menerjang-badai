package application

import (
	"context"
	"errors"
	"github.com/menerjangbadai/dashboard/internal/domain"
	"testing"
)

type taskRepoProbe struct {
	TaskRepository
	role    string
	denied  bool
	created bool
	actor   domain.Actor
}

func (r *taskRepoProbe) Membership(_ context.Context, a domain.Actor) (domain.Actor, error) {
	if r.denied {
		return a, domain.ErrForbidden
	}
	a.Role = r.role
	return a, nil
}
func (r *taskRepoProbe) GetBrand(context.Context, domain.Actor, string) (domain.Brand, error) {
	return domain.Brand{Status: "active"}, nil
}
func (r *taskRepoProbe) CreateTask(_ context.Context, a domain.Actor, in domain.Task) (domain.Task, error) {
	r.created = true
	r.actor = a
	in.ID = "new"
	in.Status = domain.TaskDraft
	return in, nil
}

func TestTaskCreateUsesCurrentMembershipNotCallerRole(t *testing.T) {
	repo := &taskRepoProbe{role: "viewer"}
	service := Tasks{Repo: repo}
	_, err := service.Create(context.Background(), domain.Actor{Role: "owner"}, domain.Task{BrandID: "d0e0ddbb-23f8-4fe7-9be7-e53dd92e9860", Title: "Brief"})
	if !errors.Is(err, domain.ErrForbidden) || repo.created {
		t.Fatal("viewer bypass", err)
	}
	repo.role = "editor"
	task, err := service.Create(context.Background(), domain.Actor{UserID: "operator"}, domain.Task{BrandID: "d0e0ddbb-23f8-4fe7-9be7-e53dd92e9860", Title: "  Brief  "})
	if err != nil || !repo.created || task.Status != domain.TaskDraft || task.Title != "Brief" || repo.actor.Role != "editor" {
		t.Fatalf("draft create: %+v %v", task, err)
	}
}
