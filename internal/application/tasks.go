package application

import (
	"context"
	"github.com/google/uuid"
	"github.com/menerjangbadai/dashboard/internal/domain"
)

type TaskRepository interface {
	Membership(context.Context, domain.Actor) (domain.Actor, error)
	GetBrand(context.Context, domain.Actor, string) (domain.Brand, error)
	ListTasks(context.Context, domain.Actor, string) ([]domain.Task, error)
	GetTask(context.Context, domain.Actor, string) (domain.Task, error)
	CreateTask(context.Context, domain.Actor, domain.Task) (domain.Task, error)
	UpdateTask(context.Context, domain.Actor, domain.Task) (domain.Task, error)
	TransitionTask(context.Context, domain.Actor, string, domain.TaskTransition) (domain.Task, error)
	TaskEvents(context.Context, domain.Actor, string) ([]domain.TaskEvent, error)
}
type Tasks struct{ Repo TaskRepository }

func (s Tasks) Create(ctx context.Context, a domain.Actor, in domain.Task) (domain.Task, error) {
	a, err := s.Repo.Membership(ctx, a)
	if err != nil {
		return domain.Task{}, err
	}
	if !canWrite(a.Role) {
		return domain.Task{}, domain.ErrForbidden
	}
	in, err = domain.NewTask(in)
	if err != nil {
		return domain.Task{}, err
	}
	brand, err := s.Repo.GetBrand(ctx, a, in.BrandID)
	if err != nil {
		return domain.Task{}, err
	}
	if brand.Status != "active" {
		return domain.Task{}, domain.ErrInvalid
	}
	return s.Repo.CreateTask(ctx, a, in)
}
func (s Tasks) List(ctx context.Context, a domain.Actor, brandID string) ([]domain.Task, error) {
	a, err := s.Repo.Membership(ctx, a)
	if err != nil {
		return nil, err
	}
	if brandID != "" {
		if _, err = uuid.Parse(brandID); err != nil {
			return nil, domain.ErrInvalid
		}
		if _, err = s.Repo.GetBrand(ctx, a, brandID); err != nil {
			return nil, err
		}
	}
	return s.Repo.ListTasks(ctx, a, brandID)
}
func (s Tasks) Get(ctx context.Context, a domain.Actor, id string) (domain.Task, error) {
	a, err := s.Repo.Membership(ctx, a)
	if err != nil {
		return domain.Task{}, err
	}
	if _, err = uuid.Parse(id); err != nil {
		return domain.Task{}, domain.ErrNotFound
	}
	return s.Repo.GetTask(ctx, a, id)
}
func (s Tasks) Update(ctx context.Context, a domain.Actor, in domain.Task) (domain.Task, error) {
	a, err := s.Repo.Membership(ctx, a)
	if err != nil {
		return domain.Task{}, err
	}
	if !canWrite(a.Role) {
		return domain.Task{}, domain.ErrForbidden
	}
	if _, err = uuid.Parse(in.ID); err != nil {
		return domain.Task{}, domain.ErrNotFound
	}
	return s.Repo.UpdateTask(ctx, a, in)
}
func (s Tasks) Transition(ctx context.Context, a domain.Actor, id string, in domain.TaskTransition) (domain.Task, error) {
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
	return s.Repo.TransitionTask(ctx, a, id, in)
}
func (s Tasks) Events(ctx context.Context, a domain.Actor, id string) ([]domain.TaskEvent, error) {
	a, err := s.Repo.Membership(ctx, a)
	if err != nil {
		return nil, err
	}
	if _, err = uuid.Parse(id); err != nil {
		return nil, domain.ErrNotFound
	}
	if _, err = s.Repo.GetTask(ctx, a, id); err != nil {
		return nil, err
	}
	return s.Repo.TaskEvents(ctx, a, id)
}
