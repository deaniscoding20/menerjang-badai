package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestDeliverableSubmissionPreservesRawAndSeparatesValidation(t *testing.T) {
	raw := ` {"schema_version":1,"kind":"research_report","title":" Audience ","sections":[{"heading":"Finding","body":"Human review required"}]} `
	result, err := NewDeliverableSubmission(DeliverableSubmission{RawOutput: raw, Version: 3})
	if err != nil || result.RawOutput != raw || result.ValidationStatus != "validated" || result.Structured == nil || result.Structured.Title != "Audience" {
		t.Fatal(result, err)
	}
	for _, raw := range []string{"ordinary unstructured output", `{"schema_version":99}`, `{"schema_version":1,"kind":"research_report","title":"Report","sections":[]}`} {
		result, err = NewDeliverableSubmission(DeliverableSubmission{RawOutput: raw, Version: 3})
		if err != nil || result.RawOutput != raw || result.ValidationStatus != "invalid" || result.Structured != nil {
			t.Fatal("invalid raw lost or promoted", result, err)
		}
	}
	literal := `{"schema_version":1,"kind":"research_report","title":"Report","sections":[{"heading":"Finding","body":"literal \\u0000 sequence"}]}`
	result, err = NewDeliverableSubmission(DeliverableSubmission{RawOutput: literal, Version: 3})
	if err != nil || result.ValidationStatus != "validated" {
		t.Fatal("literal backslash was mistaken for NUL", result, err)
	}
	escaped := `{"schema_version":1,"kind":"research_report","title":"Report","sections":[{"heading":"Finding","body":"actual escaped \u0000"}]}`
	result, err = NewDeliverableSubmission(DeliverableSubmission{RawOutput: escaped, Version: 3})
	if err != nil || result.ValidationStatus != "invalid" || result.Structured != nil {
		t.Fatal("escaped NUL must remain raw-only", result, err)
	}
	for _, raw := range []string{"", "  ", strings.Repeat("a", 128*1024+1), "nul\x00output", string([]byte{0xff})} {
		if _, err = NewDeliverableSubmission(DeliverableSubmission{RawOutput: raw, Version: 3}); !errors.Is(err, ErrInvalid) {
			t.Fatal("unsafe input", err)
		}
	}
	if _, err = NewDeliverableSubmission(DeliverableSubmission{RawOutput: "raw", Version: 0}); !errors.Is(err, ErrTaskConflict) {
		t.Fatal("missing version", err)
	}
}
