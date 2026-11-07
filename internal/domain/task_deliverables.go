package domain

import (
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"
)

// Human import does not assert that an agent executed or that content is approved.
type DeliverableSubmission struct {
	RawOutput string `json:"raw_output"`
	Version   int64  `json:"version"`
}
type ValidatedDeliverableSubmission struct {
	RawOutput        string
	ValidationStatus string
	Structured       *StructuredDeliverable
}
type DeliverableSummary struct {
	ID               string    `json:"id"`
	TaskID           string    `json:"task_id"`
	Revision         int64     `json:"revision"`
	AuthorID         string    `json:"author_id"`
	Origin           string    `json:"origin"`
	ValidationStatus string    `json:"validation_status"`
	Title            string    `json:"title"`
	Kind             string    `json:"kind"`
	CreatedAt        time.Time `json:"created_at"`
}
type DeliverableVersion struct {
	DeliverableSummary
	RawOutput  string                 `json:"raw_output"`
	Structured *StructuredDeliverable `json:"structured"`
}
type TaskDeliverables struct {
	Version int64                `json:"version"`
	Items   []DeliverableSummary `json:"items"`
}

func NewDeliverableSubmission(in DeliverableSubmission) (ValidatedDeliverableSubmission, error) {
	out := ValidatedDeliverableSubmission{}
	if in.Version < 1 {
		return out, ErrTaskConflict
	}
	if strings.TrimSpace(in.RawOutput) == "" || len(in.RawOutput) > 128*1024 || !utf8.ValidString(in.RawOutput) || strings.ContainsRune(in.RawOutput, 0) {
		return out, ErrInvalid
	}
	out.RawOutput = in.RawOutput
	out.ValidationStatus = "invalid"
	parsed, err := ParseDeliverable(json.RawMessage(in.RawOutput))
	if err == nil {
		// PostgreSQL jsonb cannot represent NUL, including escaped JSON NUL.
		hasNUL := strings.ContainsRune(parsed.Title, 0)
		for _, section := range parsed.Sections {
			hasNUL = hasNUL || strings.ContainsRune(section.Heading, 0) || strings.ContainsRune(section.Body, 0)
		}
		for _, source := range parsed.Sources {
			hasNUL = hasNUL || strings.ContainsRune(source.Label, 0) || strings.ContainsRune(source.URL, 0)
		}
		if !hasNUL {
			out.Structured = &parsed
			out.ValidationStatus = "validated"
		}
	}
	return out, nil
}
