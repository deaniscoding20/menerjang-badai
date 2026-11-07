package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestTaskDeadlinePreservesInstantAndRejectsZero(t *testing.T) {
	deadline := time.Date(2027, 1, 2, 9, 30, 0, 0, time.FixedZone("UTC+7", 7*60*60))
	task, err := NewTask(Task{BrandID: "d0e0ddbb-23f8-4fe7-9be7-e53dd92e9860", Title: "Brief", DeadlineAt: &deadline})
	if err != nil || task.DeadlineAt == nil || !task.DeadlineAt.Equal(deadline) {
		t.Fatal("deadline lost", err)
	}
	zero := time.Time{}
	task.DeadlineAt = &zero
	if !errors.Is(task.Validate(), ErrInvalid) {
		t.Fatal("invalid zero deadline accepted")
	}
}
func TestTaskCommentValidation(t *testing.T) {
	valid, err := NewTaskComment(TaskComment{Body: "  Please clarify the intended audience.  ", Kind: "clarification"})
	if err != nil || valid.Body != "Please clarify the intended audience." {
		t.Fatalf("valid clarification %+v %v", valid, err)
	}
	for _, in := range []TaskComment{{Body: " "}, {Body: strings.Repeat("x", 4001)}, {Body: "Text", Kind: "approval"}, {Body: "Text\x00"}} {
		if _, err = NewTaskComment(in); !errors.Is(err, ErrInvalid) {
			t.Fatalf("invalid comment accepted %+v %v", in, err)
		}
	}
}
