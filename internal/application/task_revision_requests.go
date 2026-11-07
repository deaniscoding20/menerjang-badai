package application

import (
	"context"
	"github.com/google/uuid"
	"github.com/menerjangbadai/dashboard/internal/domain"
)

// TaskRevisionRequestRepository stores human feedback only. Execution remains disabled.
type TaskRevisionRequestRepository interface {
	ListTaskRevisionRequests(context.Context, domain.Actor, string) (domain.TaskRevisionRequests, error)
	CreateTaskRevisionRequest(context.Context, domain.Actor, string, domain.RevisionRequestSubmission) (domain.TaskRevisionRequest, error)
}

func (s Tasks) RequestRevision(ctx context.Context, a domain.Actor, id string, in domain.RevisionRequestSubmission) (domain.TaskRevisionRequest, error) {
	a, err := s.Repo.Membership(ctx, a)
	if err != nil {
		return domain.TaskRevisionRequest{}, err
	}
	if !canWrite(a.Role) {
		return domain.TaskRevisionRequest{}, domain.ErrForbidden
	}
	if _, err = uuid.Parse(id); err != nil {
		return domain.TaskRevisionRequest{}, domain.ErrNotFound
	}
	task, err := s.Repo.GetTask(ctx, a, id)
	if err != nil {
		return domain.TaskRevisionRequest{}, err
	}
	if in.Version < 1 || task.Version != in.Version {
		return domain.TaskRevisionRequest{}, domain.ErrTaskConflict
	}
	if !task.Editable() {
		return domain.TaskRevisionRequest{}, domain.ErrInvalid
	}
	in, err = domain.NewRevisionRequestSubmission(in)
	if err != nil {
		return domain.TaskRevisionRequest{}, err
	}
	repo, ok := s.Repo.(TaskRevisionRequestRepository)
	if !ok {
		return domain.TaskRevisionRequest{}, domain.ErrTaskExecutionDisabled
	}
	return repo.CreateTaskRevisionRequest(ctx, a, id, in)
}

func (s Tasks) RevisionRequests(ctx context.Context, a domain.Actor, id string) (domain.TaskRevisionRequests, error) {
	a, err := s.Repo.Membership(ctx, a)
	if err != nil {
		return domain.TaskRevisionRequests{}, err
	}
	if _, err = uuid.Parse(id); err != nil {
		return domain.TaskRevisionRequests{}, domain.ErrNotFound
	}
	if _, err = s.Repo.GetTask(ctx, a, id); err != nil {
		return domain.TaskRevisionRequests{}, err
	}
	repo, ok := s.Repo.(TaskRevisionRequestRepository)
	if !ok {
		return domain.TaskRevisionRequests{}, domain.ErrTaskExecutionDisabled
	}
	return repo.ListTaskRevisionRequests(ctx, a, id)
}
