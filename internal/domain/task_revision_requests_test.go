package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestRevisionRequestSubmissionLimitsAndTarget(t *testing.T) {
	for _, body := range []string{"", "\u2003\n", "\u0085", strings.Repeat("🌀", 4001), "nul\x00note", string([]byte{0xff})} {
		if _, err := NewRevisionRequestSubmission(RevisionRequestSubmission{DeliverableID: revisionTargetID, Body: body, Version: 3}); !errors.Is(err, ErrInvalid) {
			t.Fatalf("body bytes=%d: %v", len(body), err)
		}
	}
	for _, body := range []string{"x", "\ufeff", strings.Repeat("🌀", 4000), "literal \ufffd"} {
		if _, err := NewRevisionRequestSubmission(RevisionRequestSubmission{DeliverableID: revisionTargetID, Body: body, Version: 3}); err != nil {
			t.Fatal(err)
		}
	}
	for _, version := range []int64{0, -1} {
		if _, err := NewRevisionRequestSubmission(RevisionRequestSubmission{DeliverableID: revisionTargetID, Body: "note", Version: version}); !errors.Is(err, ErrTaskConflict) {
			t.Fatal("positive optimistic version required", err)
		}
	}
	for _, id := range []string{"", "not-a-uuid"} {
		if _, err := NewRevisionRequestSubmission(RevisionRequestSubmission{DeliverableID: id, Body: "note", Version: 3}); !errors.Is(err, ErrNotFound) {
			t.Fatal("invalid target must be hidden", err)
		}
	}
}

const revisionTargetID = "618d5470-f095-4e3a-8a04-3bfe518b7c25"

func TestRevisionRequestSubmissionTrimsHumanBody(t *testing.T) {
	got, err := NewRevisionRequestSubmission(RevisionRequestSubmission{DeliverableID: revisionTargetID, Body: "\u2003 Please revise the evidence. \n", Version: 3})
	if err != nil || got.Body != "Please revise the evidence." || got.DeliverableID != revisionTargetID || got.Version != 3 {
		t.Fatal(got, err)
	}
}
