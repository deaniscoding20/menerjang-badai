package domain

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

type TaskStatus string

const (
	TaskDraft           TaskStatus = "draft"
	TaskReady           TaskStatus = "ready"
	TaskAssigned        TaskStatus = "assigned"
	TaskQueued          TaskStatus = "queued"
	TaskRunning         TaskStatus = "running"
	TaskWaitingInput    TaskStatus = "waiting_input"
	TaskWaitingApproval TaskStatus = "waiting_approval"
	TaskCompleted       TaskStatus = "completed"
	TaskFailed          TaskStatus = "failed"
	TaskCancelled       TaskStatus = "cancelled"
)

var ErrTaskConflict = errors.New("task version conflict")
var ErrTaskExecutionDisabled = errors.New("task execution disabled")

type Task struct {
	ID              string     `json:"id"`
	OrgID           string     `json:"-"`
	BrandID         string     `json:"brand_id"`
	RequesterID     string     `json:"requester_id"`
	Title           string     `json:"title"`
	Objective       string     `json:"objective"`
	Instructions    string     `json:"instructions"`
	Priority        string     `json:"priority"`
	Status          TaskStatus `json:"status"`
	AssignedAgentID string     `json:"agent_id"`
	Version         int64      `json:"version"`
	DeadlineAt      *time.Time `json:"deadline_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
type TaskTransition struct {
	Status  TaskStatus `json:"status"`
	Version int64      `json:"version"`
	Reason  string     `json:"reason"`
}
type TaskAssignment struct {
	AgentID string `json:"agent_id"`
	Version int64  `json:"version"`
}
type TaskEvent struct {
	ID         string     `json:"id"`
	TaskID     string     `json:"task_id"`
	ActorID    string     `json:"actor_id"`
	Event      string     `json:"event"`
	FromStatus TaskStatus `json:"from_status"`
	ToStatus   TaskStatus `json:"to_status"`
	Version    int64      `json:"version"`
	Reason     string     `json:"reason"`
	CreatedAt  time.Time  `json:"created_at"`
}

func NewTask(task Task) (Task, error) {
	// Only drafts can be created. Runtime state is not client-controlled.
	if task.Status != "" && task.Status != TaskDraft {
		return Task{}, ErrInvalid
	}
	task.Status = TaskDraft
	task.Title = strings.TrimSpace(task.Title)
	if task.Priority == "" {
		task.Priority = "normal"
	}
	if err := task.Validate(); err != nil {
		return Task{}, err
	}
	return task, nil
}
func (task Task) Validate() error {
	if task.DeadlineAt != nil && (task.DeadlineAt.IsZero() || task.DeadlineAt.Year() < 2000 || task.DeadlineAt.Year() > 9999) {
		return ErrInvalid
	}
	if _, err := uuid.Parse(task.BrandID); err != nil {
		return ErrInvalid
	}
	if !utf8.ValidString(task.Title) || strings.TrimSpace(task.Title) == "" || utf8.RuneCountInString(task.Title) > 200 || utf8.RuneCountInString(task.Objective) > 4000 || utf8.RuneCountInString(task.Instructions) > 12000 {
		return ErrInvalid
	}
	switch task.Priority {
	case "low", "normal", "high", "urgent":
	default:
		return ErrInvalid
	}
	if task.AssignedAgentID != "" {
		found := false
		for _, agent := range Agents() {
			if agent.ID == task.AssignedAgentID {
				found = true
				break
			}
		}
		if !found {
			return ErrInvalid
		}
	}
	return nil
}
func (task Task) Editable() bool {
	return task.Status == TaskDraft || task.Status == TaskReady || task.Status == TaskAssigned
}
func (task Task) ApplyAssignment(in TaskAssignment) (Task, error) {
	if in.Version < 1 || in.Version != task.Version {
		return Task{}, ErrTaskConflict
	}
	if !task.Editable() {
		return Task{}, ErrInvalid
	}
	agentChanged := task.AssignedAgentID != in.AgentID
	task.AssignedAgentID = in.AgentID
	if err := task.Validate(); err != nil {
		return Task{}, err
	}
	if agentChanged && task.Status != TaskDraft {
		if in.AgentID == "" {
			task.Status = TaskReady
		} else {
			task.Status = TaskAssigned
		}
	}
	return task, nil
}
func (task Task) ApplyTransition(in TaskTransition) (Task, error) {
	if in.Version < 1 || in.Version != task.Version {
		return Task{}, ErrTaskConflict
	}
	if utf8.RuneCountInString(in.Reason) > 2000 {
		return Task{}, ErrInvalid
	}
	if !task.Editable() {
		return Task{}, ErrInvalid
	}
	switch in.Status {
	case TaskQueued, TaskRunning, TaskWaitingInput, TaskWaitingApproval, TaskCompleted, TaskFailed:
		return Task{}, ErrTaskExecutionDisabled
	case TaskReady:
		if task.Status != TaskDraft && task.Status != TaskAssigned {
			return Task{}, ErrInvalid
		}
	case TaskAssigned:
		if task.Status != TaskReady || task.AssignedAgentID == "" {
			return Task{}, ErrInvalid
		}
	case TaskCancelled:
	default:
		return Task{}, ErrInvalid
	}
	task.Status = in.Status
	return task, nil
}
