package application

import (
	"context"
	"github.com/google/uuid"
	"github.com/menerjangbadai/dashboard/internal/domain"
)

type TaskDeliverableRepository interface {
	ListTaskDeliverables(context.Context, domain.Actor, string) (domain.TaskDeliverables, error)
	GetTaskDeliverable(context.Context, domain.Actor, string, string) (domain.DeliverableVersion, error)
	SubmitTaskDeliverable(context.Context, domain.Actor, string, domain.DeliverableSubmission) (domain.DeliverableVersion, error)
}

func (s Tasks) SubmitDeliverable(ctx context.Context, a domain.Actor, id string, in domain.DeliverableSubmission) (domain.DeliverableVersion, error) {
	a, err := s.Repo.Membership(ctx, a)
	if err != nil {
		return domain.DeliverableVersion{}, err
	}
	if !canWrite(a.Role) {
		return domain.DeliverableVersion{}, domain.ErrForbidden
	}
	if _, err = uuid.Parse(id); err != nil {
		return domain.DeliverableVersion{}, domain.ErrNotFound
	}
	task, err := s.Repo.GetTask(ctx, a, id)
	if err != nil {
		return domain.DeliverableVersion{}, err
	}
	if in.Version < 1 || task.Version != in.Version {
		return domain.DeliverableVersion{}, domain.ErrTaskConflict
	}
	if !task.Editable() {
		return domain.DeliverableVersion{}, domain.ErrInvalid
	}
	if _, err = domain.NewDeliverableSubmission(in); err != nil {
		return domain.DeliverableVersion{}, err
	}
	repo, ok := s.Repo.(TaskDeliverableRepository)
	if !ok {
		return domain.DeliverableVersion{}, domain.ErrTaskExecutionDisabled
	}
	return repo.SubmitTaskDeliverable(ctx, a, id, in)
}
func (s Tasks) Deliverables(ctx context.Context, a domain.Actor, id string) (domain.TaskDeliverables, error) {
	a, err := s.Repo.Membership(ctx, a)
	if err != nil {
		return domain.TaskDeliverables{}, err
	}
	if _, err = uuid.Parse(id); err != nil {
		return domain.TaskDeliverables{}, domain.ErrNotFound
	}
	if _, err = s.Repo.GetTask(ctx, a, id); err != nil {
		return domain.TaskDeliverables{}, err
	}
	repo, ok := s.Repo.(TaskDeliverableRepository)
	if !ok {
		return domain.TaskDeliverables{}, domain.ErrTaskExecutionDisabled
	}
	return repo.ListTaskDeliverables(ctx, a, id)
}
func (s Tasks) Deliverable(ctx context.Context, a domain.Actor, id string, versionID string) (domain.DeliverableVersion, error) {
	a, err := s.Repo.Membership(ctx, a)
	if err != nil {
		return domain.DeliverableVersion{}, err
	}
	if _, err = uuid.Parse(id); err != nil {
		return domain.DeliverableVersion{}, domain.ErrNotFound
	}
	if _, err = uuid.Parse(versionID); err != nil {
		return domain.DeliverableVersion{}, domain.ErrNotFound
	}
	if _, err = s.Repo.GetTask(ctx, a, id); err != nil {
		return domain.DeliverableVersion{}, err
	}
	repo, ok := s.Repo.(TaskDeliverableRepository)
	if !ok {
		return domain.DeliverableVersion{}, domain.ErrTaskExecutionDisabled
	}
	return repo.GetTaskDeliverable(ctx, a, id, versionID)
}
