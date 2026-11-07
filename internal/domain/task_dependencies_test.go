package domain

import (
	"errors"
	"testing"
)

func TestDependencyChangesRequireCurrentEditableTask(t *testing.T) {
	task := Task{ID: "dbf244dd-5907-4308-bdb4-6fe2d3d58cd8", Status: TaskReady, Version: 3}
	valid := TaskDependencyChange{DependencyID: "89b899e1-5922-4053-9417-7eaf90d170f7", Version: 3}
	if err := task.ValidateDependencyChange(valid); err != nil {
		t.Fatal(err)
	}
	if task.Status != TaskReady || task.Version != 3 {
		t.Fatal("validation mutated lifecycle")
	}
	stale := valid
	stale.Version = 2
	if err := task.ValidateDependencyChange(stale); !errors.Is(err, ErrTaskConflict) {
		t.Fatal("stale dependency change", err)
	}
	self := valid
	self.DependencyID = "DBF244DD-5907-4308-BDB4-6FE2D3D58CD8"
	if err := task.ValidateDependencyChange(self); !errors.Is(err, ErrInvalid) {
		t.Fatal("self dependency", err)
	}
	malformed := valid
	malformed.DependencyID = "not-a-task"
	if err := task.ValidateDependencyChange(malformed); !errors.Is(err, ErrInvalid) {
		t.Fatal("invalid dependency", err)
	}
	for _, status := range []TaskStatus{TaskQueued, TaskRunning, TaskWaitingInput, TaskWaitingApproval, TaskCompleted, TaskFailed, TaskCancelled} {
		task.Status = status
		if err := task.ValidateDependencyChange(valid); !errors.Is(err, ErrInvalid) {
			t.Fatalf("dependency edit in %s: %v", status, err)
		}
	}
}
