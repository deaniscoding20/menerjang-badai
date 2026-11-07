package domain

import (
	"errors"
	"github.com/google/uuid"
)

var ErrDependencyCycle = errors.New("task dependency cycle")

type TaskDependency struct {
	TaskID string     `json:"task_id"`
	Title  string     `json:"title"`
	Status TaskStatus `json:"status"`
}

type TaskDependencies struct {
	Version int64            `json:"version"`
	Items   []TaskDependency `json:"items"`
}

type TaskDependencyChange struct {
	DependencyID string `json:"dependency_id"`
	Version      int64  `json:"version"`
}

// Dependency changes edit preparation metadata, not execution or task status.
func (task Task) ValidateDependencyChange(in TaskDependencyChange) error {
	if in.Version < 1 || task.Version != in.Version {
		return ErrTaskConflict
	}
	if !task.Editable() {
		return ErrInvalid
	}
	source, err := uuid.Parse(task.ID)
	if err != nil {
		return ErrInvalid
	}
	target, err := uuid.Parse(in.DependencyID)
	if err != nil || target == uuid.Nil || source == target {
		return ErrInvalid
	}
	return nil
}
