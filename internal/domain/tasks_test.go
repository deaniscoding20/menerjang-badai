package domain

import (
	"errors"
	"testing"
	"time"
)

func TestTaskAssignmentPreservesReadyOnUnchangedAgent(t *testing.T) {
	deadline := time.Date(2027, 1, 2, 9, 30, 0, 0, time.UTC)
	task := Task{BrandID: "d0e0ddbb-23f8-4fe7-9be7-e53dd92e9860", Title: "Brief", Priority: "normal", Status: TaskReady, AssignedAgentID: Agents()[4].ID, Version: 3, DeadlineAt: &deadline}
	for _, edit := range []string{"brief", "change deadline", "clear deadline"} {
		t.Run(edit, func(t *testing.T) {
			edited := task
			switch edit {
			case "brief":
				edited.Title = "Updated brief"
				edited.Objective = "Updated objective"
				edited.Instructions = "Updated instructions"
			case "change deadline":
				later := deadline.Add(time.Hour)
				edited.DeadlineAt = &later
			case "clear deadline":
				edited.DeadlineAt = nil
			}
			got, err := edited.ApplyAssignment(TaskAssignment{AgentID: edited.AssignedAgentID, Version: edited.Version})
			if err != nil || got.Status != TaskReady || got != edited {
				t.Fatalf("metadata update changed ready task: %+v, %v", got, err)
			}
		})
	}
	// A ready task with a configured agent still supports an explicit transition.
	assigned, err := task.ApplyTransition(TaskTransition{Status: TaskAssigned, Version: task.Version})
	if err != nil || assigned.Status != TaskAssigned {
		t.Fatalf("explicit assignment: %+v, %v", assigned, err)
	}
	for _, agentID := range []string{Agents()[0].ID, ""} {
		got, err := task.ApplyAssignment(TaskAssignment{AgentID: agentID, Version: task.Version})
		want := TaskAssigned
		if agentID == "" {
			want = TaskReady
		}
		if err != nil || got.Status != want || got.AssignedAgentID != agentID {
			t.Fatalf("changed assignment %q: %+v, %v", agentID, got, err)
		}
	}
}

func TestTaskUnchangedAssignmentStillChecksGuards(t *testing.T) {
	task := Task{BrandID: "d0e0ddbb-23f8-4fe7-9be7-e53dd92e9860", Title: "Brief", Priority: "normal", Status: TaskReady, AssignedAgentID: Agents()[4].ID, Version: 3}
	if _, err := task.ApplyAssignment(TaskAssignment{AgentID: task.AssignedAgentID, Version: 2}); !errors.Is(err, ErrTaskConflict) {
		t.Fatalf("stale unchanged assignment: %v", err)
	}
	for _, invalid := range []Task{
		{BrandID: task.BrandID, Title: "", Priority: task.Priority, Status: task.Status, AssignedAgentID: task.AssignedAgentID, Version: task.Version},
		{BrandID: task.BrandID, Title: task.Title, Priority: task.Priority, Status: TaskCancelled, AssignedAgentID: task.AssignedAgentID, Version: task.Version},
	} {
		if _, err := invalid.ApplyAssignment(TaskAssignment{AgentID: invalid.AssignedAgentID, Version: invalid.Version}); !errors.Is(err, ErrInvalid) {
			t.Fatalf("unchanged assignment bypassed validation/editability: %v", err)
		}
	}
}

func TestTaskDevelopmentLifecycle(t *testing.T) {
	task, err := NewTask(Task{BrandID: "d0e0ddbb-23f8-4fe7-9be7-e53dd92e9860", Title: "  Campaign brief  "})
	if err != nil || task.Status != TaskDraft || task.Title != "Campaign brief" {
		t.Fatalf("new draft: %+v, %v", task, err)
	}
	task.Version = 1 // Persistence assigns the initial version.
	task, err = task.ApplyTransition(TaskTransition{Status: TaskReady, Version: 1})
	if err != nil {
		t.Fatal(err)
	}
	// A direct specialist assignment does not have to pass through Claudia.
	task, err = task.ApplyAssignment(TaskAssignment{AgentID: Agents()[4].ID, Version: 1})
	if err != nil || task.Status != TaskAssigned || task.AssignedAgentID != Agents()[4].ID {
		t.Fatalf("direct assignment: %+v, %v", task, err)
	}
	for _, status := range []TaskStatus{TaskQueued, TaskRunning, TaskWaitingInput, TaskWaitingApproval, TaskCompleted, TaskFailed} {
		if _, err = task.ApplyTransition(TaskTransition{Status: status, Version: 1}); !errors.Is(err, ErrTaskExecutionDisabled) {
			t.Fatalf("unverified transition %q: %v", status, err)
		}
	}
	task, err = task.ApplyTransition(TaskTransition{Status: TaskCancelled, Version: 1, Reason: "No longer needed"})
	if err != nil || task.Status != TaskCancelled {
		t.Fatalf("cancel: %+v, %v", task, err)
	}
	if _, err = task.ApplyTransition(TaskTransition{Status: TaskReady, Version: 1}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("terminal task reopened: %v", err)
	}
}
