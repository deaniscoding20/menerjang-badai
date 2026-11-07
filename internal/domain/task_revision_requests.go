package domain

import (
	"github.com/google/uuid"
	"strings"
	"time"
	"unicode/utf8"
)

// RevisionRequestSubmission is a human note about an immutable output version.
// It is not an instruction to dispatch, approve, retry or resume execution.
type RevisionRequestSubmission struct {
	DeliverableID string `json:"deliverable_id"`
	Body          string `json:"body"`
	Version       int64  `json:"version"`
}

type TaskRevisionRequest struct {
	ID                  string    `json:"id"`
	TaskID              string    `json:"task_id"`
	DeliverableID       string    `json:"deliverable_id"`
	DeliverableRevision int64     `json:"deliverable_revision"`
	AuthorID            string    `json:"author_id"`
	Body                string    `json:"body"`
	CreatedAt           time.Time `json:"created_at"`
}

type TaskRevisionRequests struct {
	Version int64                 `json:"version"`
	Items   []TaskRevisionRequest `json:"items"`
}

func NewRevisionRequestSubmission(in RevisionRequestSubmission) (RevisionRequestSubmission, error) {
	if in.Version < 1 {
		return RevisionRequestSubmission{}, ErrTaskConflict
	}
	if _, err := uuid.Parse(in.DeliverableID); err != nil {
		return RevisionRequestSubmission{}, ErrNotFound
	}
	in.Body = strings.TrimSpace(in.Body)
	if in.Body == "" || !utf8.ValidString(in.Body) || utf8.RuneCountInString(in.Body) > 4000 || strings.ContainsRune(in.Body, 0) {
		return RevisionRequestSubmission{}, ErrInvalid
	}
	return in, nil
}
