package application

import (
	"context"
	"errors"
	"github.com/menerjangbadai/dashboard/internal/domain"
	"testing"
)

type deliverableRepo struct {
	TaskRepository
	role   string
	hidden bool
	status domain.TaskStatus
	saved  bool
}

func (r *deliverableRepo) Membership(_ context.Context, a domain.Actor) (domain.Actor, error) {
	a.Role = r.role
	return a, nil
}
func (r *deliverableRepo) GetTask(_ context.Context, _ domain.Actor, id string) (domain.Task, error) {
	if r.hidden {
		return domain.Task{}, domain.ErrNotFound
	}
	return domain.Task{ID: id, Version: 3, Status: r.status}, nil
}
func (r *deliverableRepo) ListTaskDeliverables(context.Context, domain.Actor, string) (domain.TaskDeliverables, error) {
	return domain.TaskDeliverables{Version: 3, Items: []domain.DeliverableSummary{}}, nil
}
func (r *deliverableRepo) GetTaskDeliverable(context.Context, domain.Actor, string, string) (domain.DeliverableVersion, error) {
	return domain.DeliverableVersion{RawOutput: "original"}, nil
}
func (r *deliverableRepo) SubmitTaskDeliverable(_ context.Context, a domain.Actor, id string, in domain.DeliverableSubmission) (domain.DeliverableVersion, error) {
	r.saved = true
	return domain.DeliverableVersion{DeliverableSummary: domain.DeliverableSummary{TaskID: id, AuthorID: a.UserID, Revision: 1, Origin: "human_import", ValidationStatus: "invalid"}, RawOutput: in.RawOutput}, nil
}
func TestDeliverableServiceCurrentAuthorityVersionAndLifecycle(t *testing.T) {
	repo := &deliverableRepo{role: "viewer", status: domain.TaskReady}
	service := Tasks{Repo: repo}
	id := "dbf244dd-5907-4308-bdb4-6fe2d3d58cd8"
	in := domain.DeliverableSubmission{RawOutput: "raw output", Version: 3}
	if _, err := service.SubmitDeliverable(context.Background(), domain.Actor{Role: "owner"}, id, in); !errors.Is(err, domain.ErrForbidden) || repo.saved {
		t.Fatal("claimed role", err)
	}
	repo.role = "editor"
	repo.hidden = true
	if _, err := service.SubmitDeliverable(context.Background(), domain.Actor{}, id, in); !errors.Is(err, domain.ErrNotFound) || repo.saved {
		t.Fatal("hidden task", err)
	}
	repo.hidden = false
	repo.status = domain.TaskCompleted
	if _, err := service.SubmitDeliverable(context.Background(), domain.Actor{}, id, in); !errors.Is(err, domain.ErrInvalid) || repo.saved {
		t.Fatal("terminal mutation", err)
	}
	repo.status = domain.TaskReady
	in.Version = 2
	if _, err := service.SubmitDeliverable(context.Background(), domain.Actor{}, id, in); !errors.Is(err, domain.ErrTaskConflict) || repo.saved {
		t.Fatal("stale version", err)
	}
	in.Version = 3
	got, err := service.SubmitDeliverable(context.Background(), domain.Actor{UserID: "session-author"}, id, in)
	if err != nil || !repo.saved || got.RawOutput != in.RawOutput || got.Origin != "human_import" || got.AuthorID != "session-author" {
		t.Fatal(got, err)
	}
	repo.role = "viewer"
	if _, err = service.Deliverables(context.Background(), domain.Actor{}, id); err != nil {
		t.Fatal("viewer list", err)
	}
	if _, err = service.Deliverable(context.Background(), domain.Actor{}, id, "bad"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("bad revision id", err)
	}
}
