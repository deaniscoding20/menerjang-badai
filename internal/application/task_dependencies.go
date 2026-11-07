package application

import (
	"context"
	"github.com/google/uuid"
	"github.com/menerjangbadai/dashboard/internal/domain"
)

type TaskDependencyRepository interface {
	ListTaskDependencies(context.Context, domain.Actor, string) (domain.TaskDependencies, error)
	ChangeTaskDependency(context.Context, domain.Actor, string, domain.TaskDependencyChange, bool) (domain.Task, error)
}

func (s Tasks) Dependencies(ctx context.Context, a domain.Actor, id string) (domain.TaskDependencies, error) {
	a, err := s.Repo.Membership(ctx, a)
	if err != nil {
		return domain.TaskDependencies{}, err
	}
	if _, err = uuid.Parse(id); err != nil {
		return domain.TaskDependencies{}, domain.ErrNotFound
	}
	if _, err = s.Repo.GetTask(ctx, a, id); err != nil {
		return domain.TaskDependencies{}, err
	}
	repo, ok := s.Repo.(TaskDependencyRepository)
	if !ok {
		return domain.TaskDependencies{}, domain.ErrTaskExecutionDisabled
	}
	return repo.ListTaskDependencies(ctx, a, id)
}

func (s Tasks) ChangeDependency(ctx context.Context, a domain.Actor, id string, in domain.TaskDependencyChange, add bool) (domain.Task, error) {
	a, err := s.Repo.Membership(ctx, a)
	if err != nil {
		return domain.Task{}, err
	}
	if !canWrite(a.Role) {
		return domain.Task{}, domain.ErrForbidden
	}
	if _, err = uuid.Parse(id); err != nil {
		return domain.Task{}, domain.ErrNotFound
	}
	task, err := s.Repo.GetTask(ctx, a, id)
	if err != nil {
		return domain.Task{}, err
	}
	if err = task.ValidateDependencyChange(in); err != nil {
		return domain.Task{}, err
	}
	target, err := s.Repo.GetTask(ctx, a, in.DependencyID)
	if err != nil {
		return domain.Task{}, err
	}
	if target.BrandID != task.BrandID {
		return domain.Task{}, domain.ErrNotFound
	}
	repo, ok := s.Repo.(TaskDependencyRepository)
	if !ok {
		return domain.Task{}, domain.ErrTaskExecutionDisabled
	}
	return repo.ChangeTaskDependency(ctx, a, id, in, add)
}
