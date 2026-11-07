package application

import (
	"context"
	"errors"
	"github.com/menerjangbadai/dashboard/internal/domain"
	"testing"
	"time"
)

const revisionTaskID = "dbf244dd-5907-4308-bdb4-6fe2d3d58cd8"
const revisionDeliverableID = "618d5470-f095-4e3a-8a04-3bfe518b7c25"

type revisionRequestRepo struct {
	TaskRepository
	role      string
	task      domain.Task
	hidden    bool
	saved     int
	listed    int
	targetErr error
	in        domain.RevisionRequestSubmission
}

func (r *revisionRequestRepo) Membership(_ context.Context, a domain.Actor) (domain.Actor, error) {
	a.Role = r.role
	return a, nil
}
func (r *revisionRequestRepo) GetTask(context.Context, domain.Actor, string) (domain.Task, error) {
	if r.hidden {
		return domain.Task{}, domain.ErrNotFound
	}
	return r.task, nil
}
func (r *revisionRequestRepo) ListTaskRevisionRequests(context.Context, domain.Actor, string) (domain.TaskRevisionRequests, error) {
	r.listed++
	return domain.TaskRevisionRequests{Version: r.task.Version, Items: []domain.TaskRevisionRequest{}}, nil
}
func (r *revisionRequestRepo) CreateTaskRevisionRequest(_ context.Context, a domain.Actor, id string, in domain.RevisionRequestSubmission) (domain.TaskRevisionRequest, error) {
	if r.targetErr != nil {
		return domain.TaskRevisionRequest{}, r.targetErr
	}
	r.saved++
	r.in = in
	r.task.Version++
	return domain.TaskRevisionRequest{ID: revisionTaskID, TaskID: id, DeliverableID: in.DeliverableID, DeliverableRevision: 1, AuthorID: a.UserID, Body: in.Body, CreatedAt: time.Now()}, nil
}
func TestRevisionRequestServiceRecordsNormalizedHumanNote(t *testing.T) {
	repo := &revisionRequestRepo{role: "editor", task: domain.Task{ID: revisionTaskID, Version: 3, Status: domain.TaskReady}}
	got, err := (Tasks{Repo: repo}).RequestRevision(context.Background(), domain.Actor{UserID: "session-author"}, revisionTaskID, domain.RevisionRequestSubmission{DeliverableID: revisionDeliverableID, Body: "  Clarify evidence. \n", Version: 3})
	if err != nil || repo.saved != 1 || repo.in.Body != "Clarify evidence." || got.Body != repo.in.Body || got.AuthorID != "session-author" || got.DeliverableRevision != 1 || repo.task.Status != domain.TaskReady || repo.task.Version != 4 {
		t.Fatal(got, err, repo)
	}
}
func TestRevisionRequestServiceCurrentAuthorityParentAndEditableState(t *testing.T) {
	for _, role := range []string{"viewer", "approver", "analyst"} {
		repo := &revisionRequestRepo{role: role, task: domain.Task{Version: 3, Status: domain.TaskReady}}
		_, err := (Tasks{Repo: repo}).RequestRevision(context.Background(), domain.Actor{Role: "owner"}, revisionTaskID, domain.RevisionRequestSubmission{DeliverableID: revisionDeliverableID, Body: "note", Version: 3})
		if !errors.Is(err, domain.ErrForbidden) || repo.saved != 0 {
			t.Fatal(role, err, repo.saved)
		}
	}
	for _, status := range []domain.TaskStatus{domain.TaskQueued, domain.TaskRunning, domain.TaskWaitingInput, domain.TaskWaitingApproval, domain.TaskCompleted, domain.TaskFailed, domain.TaskCancelled} {
		repo := &revisionRequestRepo{role: "editor", task: domain.Task{Version: 3, Status: status}}
		_, err := (Tasks{Repo: repo}).RequestRevision(context.Background(), domain.Actor{}, revisionTaskID, domain.RevisionRequestSubmission{DeliverableID: revisionDeliverableID, Body: "note", Version: 3})
		if !errors.Is(err, domain.ErrInvalid) || repo.saved != 0 {
			t.Fatal(status, err, repo.saved)
		}
	}
	for _, role := range []string{"owner", "administrator", "marketing_manager", "editor"} {
		for _, status := range []domain.TaskStatus{domain.TaskDraft, domain.TaskReady, domain.TaskAssigned} {
			repo := &revisionRequestRepo{role: role, task: domain.Task{Version: 3, Status: status}}
			_, err := (Tasks{Repo: repo}).RequestRevision(context.Background(), domain.Actor{}, revisionTaskID, domain.RevisionRequestSubmission{DeliverableID: revisionDeliverableID, Body: "note", Version: 3})
			if err != nil || repo.saved != 1 || repo.task.Status != status {
				t.Fatal(role, status, err)
			}
		}
	}
	repo := &revisionRequestRepo{role: "editor", hidden: true, task: domain.Task{Version: 3, Status: domain.TaskReady}}
	service := Tasks{Repo: repo}
	in := domain.RevisionRequestSubmission{DeliverableID: revisionDeliverableID, Body: "note", Version: 3}
	if _, err := service.RequestRevision(context.Background(), domain.Actor{}, revisionTaskID, in); !errors.Is(err, domain.ErrNotFound) || repo.saved != 0 {
		t.Fatal("hidden parent", err)
	}
	repo.hidden = false
	if _, err := service.RequestRevision(context.Background(), domain.Actor{}, "bad", in); !errors.Is(err, domain.ErrNotFound) || repo.saved != 0 {
		t.Fatal("invalid parent", err)
	}
	repo.targetErr = domain.ErrNotFound
	if _, err := service.RequestRevision(context.Background(), domain.Actor{}, revisionTaskID, in); !errors.Is(err, domain.ErrNotFound) || repo.saved != 0 {
		t.Fatal("wrong target", err)
	}
}

func TestRevisionRequestServiceListUsesCurrentMembershipAndVisibleParent(t *testing.T) {
	repo := &revisionRequestRepo{role: "viewer", task: domain.Task{Version: 3, Status: domain.TaskCompleted}}
	service := Tasks{Repo: repo}
	got, err := service.RevisionRequests(context.Background(), domain.Actor{Role: "owner"}, revisionTaskID)
	if err != nil || repo.listed != 1 || got.Version != 3 || got.Items == nil {
		t.Fatal(got, err, repo.listed)
	}
	if _, err = service.RevisionRequests(context.Background(), domain.Actor{}, "bad"); !errors.Is(err, domain.ErrNotFound) || repo.listed != 1 {
		t.Fatal("invalid parent", err)
	}
	repo.hidden = true
	if _, err = service.RevisionRequests(context.Background(), domain.Actor{}, revisionTaskID); !errors.Is(err, domain.ErrNotFound) || repo.listed != 1 {
		t.Fatal("hidden parent", err)
	}
}

func TestRevisionRequestServiceRejectsStaleWithoutRetry(t *testing.T) {
	repo := &revisionRequestRepo{role: "editor", task: domain.Task{ID: revisionTaskID, Version: 3, Status: domain.TaskReady}}
	_, err := (Tasks{Repo: repo}).RequestRevision(context.Background(), domain.Actor{}, revisionTaskID, domain.RevisionRequestSubmission{DeliverableID: revisionDeliverableID, Body: "same note", Version: 2})
	if !errors.Is(err, domain.ErrTaskConflict) || repo.saved != 0 {
		t.Fatal(err, repo.saved)
	}
}
