package domain

import (
	"strings"
	"time"
	"unicode/utf8"
)

type TaskComment struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	AuthorID  string    `json:"author_id"`
	Kind      string    `json:"kind"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

// Comments are human notes only; clarification does not resume an agent or approve output.
func NewTaskComment(comment TaskComment) (TaskComment, error) {
	comment.Body = strings.TrimSpace(comment.Body)
	if comment.Kind == "" {
		comment.Kind = "comment"
	}
	if (comment.Kind != "comment" && comment.Kind != "clarification") || comment.Body == "" || !utf8.ValidString(comment.Body) || utf8.RuneCountInString(comment.Body) > 4000 || strings.ContainsRune(comment.Body, 0) {
		return TaskComment{}, ErrInvalid
	}
	return comment, nil
}
