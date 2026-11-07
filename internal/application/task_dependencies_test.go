package application

import (
	"context"
	"errors"
	"github.com/menerjangbadai/dashboard/internal/domain"
	"testing"
)

const dependencySource = "dbf244dd-5907-4308-bdb4-6fe2d3d58cd8"
const dependencyTarget = "89b899e1-5922-4053-9417-7eaf90d170f7"

type dependencyRepo struct {
	TaskRepository
	role           string
	hidden         bool
	otherBrand     bool
	changed        bool
	failMembership bool
}

func (r *dependencyRepo) Membership(_ context.Context, a domain.Actor) (domain.Actor, error) {
	if r.failMembership {
		return domain.Actor{}, domain.ErrUnauthorized
	}
	a.Role = r.role
	return a, nil
}
func (r *dependencyRepo) GetTask(_ context.Context, _ domain.Actor, id string) (domain.Task, error) {
	if r.hidden {
		return domain.Task{}, domain.ErrNotFound
	}
	brand := "d0e0ddbb-23f8-4fe7-9be7-e53dd92e9860"
	if id == dependencyTarget && r.otherBrand {
		brand = "e5c33d94-fda1-47ad-80c4-8b91915c4413"
	}
	return domain.Task{ID: id, BrandID: brand, Status: domain.TaskReady, Version: 3}, nil
}
func (r *dependencyRepo) ListTaskDependencies(context.Context, domain.Actor, string) (domain.TaskDependencies, error) {
	return domain.TaskDependencies{Version: 3, Items: []domain.TaskDependency{}}, nil
}
func (r *dependencyRepo) ChangeTaskDependency(_ context.Context, _ domain.Actor, id string, _ domain.TaskDependencyChange, _ bool) (domain.Task, error) {
	r.changed = true
	return domain.Task{ID: id, Version: 4, Status: domain.TaskReady}, nil
}
func TestDependencyServiceRequiresCurrentWriteRoleAndSameVisibleBrand(t *testing.T) {
	repo := &dependencyRepo{role: "viewer"}
	service := Tasks{Repo: repo}
	in := domain.TaskDependencyChange{DependencyID: dependencyTarget, Version: 3}
	for _, role := range []string{"viewer", "analyst", "approver"} {
		repo.role = role
		if _, err := service.ChangeDependency(context.Background(), domain.Actor{Role: "owner"}, dependencySource, in, true); !errors.Is(err, domain.ErrForbidden) || repo.changed {
			t.Fatal(role, err)
		}
	}
	repo.role = "editor"
	repo.hidden = true
	if _, err := service.ChangeDependency(context.Background(), domain.Actor{}, dependencySource, in, true); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal(err)
	}
	repo.hidden = false
	repo.otherBrand = true
	if _, err := service.ChangeDependency(context.Background(), domain.Actor{}, dependencySource, in, true); !errors.Is(err, domain.ErrNotFound) || repo.changed {
		t.Fatal("cross brand", err)
	}
	repo.otherBrand = false
	got, err := service.ChangeDependency(context.Background(), domain.Actor{}, dependencySource, in, true)
	if err != nil || got.Version != 4 || got.Status != domain.TaskReady || !repo.changed {
		t.Fatal(got, err)
	}
	repo.role = "viewer"
	if _, err := service.Dependencies(context.Background(), domain.Actor{}, dependencySource); err != nil {
		t.Fatal("viewer read", err)
	}
	repo.failMembership = true
	if _, err := service.Dependencies(context.Background(), domain.Actor{}, dependencySource); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatal("inactive membership", err)
	}
}
