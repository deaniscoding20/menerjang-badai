package application

import (
	"context"
	"github.com/google/uuid"
	"github.com/menerjangbadai/dashboard/internal/domain"
)

type TaskCommentRepository interface {
	ListTaskComments(context.Context, domain.Actor, string) ([]domain.TaskComment, error)
	AddTaskComment(context.Context, domain.Actor, string, domain.TaskComment) (domain.TaskComment, error)
}

func (s Tasks) AddComment(ctx context.Context, a domain.Actor, id string, in domain.TaskComment) (domain.TaskComment, error) {
	a, err := s.Repo.Membership(ctx, a)
	if err != nil {
		return domain.TaskComment{}, err
	}
	if !canWrite(a.Role) {
		return domain.TaskComment{}, domain.ErrForbidden
	}
	if _, err = uuid.Parse(id); err != nil {
		return domain.TaskComment{}, domain.ErrNotFound
	}
	in, err = domain.NewTaskComment(in)
	if err != nil {
		return domain.TaskComment{}, err
	}
	if _, err = s.Repo.GetTask(ctx, a, id); err != nil {
		return domain.TaskComment{}, err
	}
	repo, ok := s.Repo.(TaskCommentRepository)
	if !ok {
		return domain.TaskComment{}, domain.ErrTaskExecutionDisabled
	}
	return repo.AddTaskComment(ctx, a, id, in)
}
func (s Tasks) Comments(ctx context.Context, a domain.Actor, id string) ([]domain.TaskComment, error) {
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
	repo, ok := s.Repo.(TaskCommentRepository)
	if !ok {
		return nil, domain.ErrTaskExecutionDisabled
	}
	return repo.ListTaskComments(ctx, a, id)
}
