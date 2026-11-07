package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestDeliverableRejectsArbitraryTextAndUnknownSchema(t *testing.T) {
	for _, raw := range []string{`not structured output`, `{"schema_version":9,"kind":"research_report","title":"Report","sections":[{"heading":"Finding","body":"Evidence"}]}`, `{"schema_version":1,"kind":"research_report","title":"Report","sections":[]}`, `{"schema_version":1,"kind":"research_report","title":"Report","sections":[{"heading":"Finding","body":"Evidence"}],"surprise":true}`} {
		if _, err := ParseDeliverable(json.RawMessage(raw)); !errors.Is(err, ErrInvalid) {
			t.Fatalf("malformed deliverable accepted: %v", err)
		}
	}
}

func TestDeliverableKindsAndSafeReferences(t *testing.T) {
	for _, kind := range []string{"research_report", "marketing_strategy", "creative_brief", "caption_package", "design_specification", "campaign_plan", "performance_report"} {
		raw, _ := json.Marshal(map[string]any{"schema_version": 1, "kind": kind, "title": "Campaign output", "sections": []map[string]string{{"heading": "Overview", "body": "Human review still required."}}, "sources": []map[string]string{{"label": "Public reference", "url": "https://example.com/report"}}})
		out, err := ParseDeliverable(raw)
		if err != nil || out.Kind != kind {
			t.Fatalf("valid %s: %v", kind, err)
		}
	}
	for _, source := range []string{"javascript:alert(1)", "https://username:password@example.com/report", "https://example.com/?token=secret", "file:///etc/passwd"} {
		raw, _ := json.Marshal(map[string]any{"schema_version": 1, "kind": "research_report", "title": "Report", "sections": []map[string]string{{"heading": "Overview", "body": "Text"}}, "sources": []map[string]string{{"label": "Reference", "url": source}}})
		if _, err := ParseDeliverable(raw); !errors.Is(err, ErrInvalid) {
			t.Fatalf("unsafe source accepted: %v", err)
		}
	}
}

func TestDeliverableTextResponseContract(t *testing.T) {
	for _, field := range []string{"title", "heading", "body", "label"} {
		for _, tc := range []struct {
			name  string
			text  string
			valid bool
		}{
			{"U+FEFF", "\uFEFF", true},
			{"U+0085", "\u0085", false},
			{"all Go whitespace", " \t\n\r\v\f\u0085\u00A0\u1680\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007\u2008\u2009\u200A\u2028\u2029\u202F\u205F\u3000", false},
		} {
			t.Run(field+"/"+tc.name, func(t *testing.T) {
				input := StructuredDeliverable{
					SchemaVersion: 1, Kind: "research_report", Title: "Report",
					Sections: []DeliverableSection{{Heading: "Finding", Body: "Evidence"}},
					Sources:  []DeliverableSource{{Label: "Public reference", URL: "https://example.com/"}},
				}
				switch field {
				case "title":
					input.Title = tc.text
				case "heading":
					input.Sections[0].Heading = tc.text
				case "body":
					input.Sections[0].Body = tc.text
				case "label":
					input.Sources[0].Label = tc.text
				}
				raw, err := json.Marshal(input)
				if err != nil {
					t.Fatal(err)
				}
				out, err := ParseDeliverable(raw)
				if !tc.valid {
					if !errors.Is(err, ErrInvalid) {
						t.Fatalf("blank text accepted: %v", err)
					}
					return
				}
				if err != nil {
					t.Fatalf("Go-nonblank text rejected: %v", err)
				}
				if out.Title != input.Title || out.Sections[0] != input.Sections[0] || out.Sources[0] != input.Sources[0] {
					t.Fatalf("nonblank text was normalized: %#v", out)
				}
			})
		}
	}
}

func TestDeliverableSourceResponseContract(t *testing.T) {
	for _, tc := range []struct {
		name   string
		source string
		valid  bool
	}{
		{"empty query separator", "https://example.com/?", true},
		{"empty query before fragment", "https://example.com/?#reference", true},
		{"question mark inside fragment", "https://example.com/#reference?token=public-text", true},
		{"UTF-8 byte limit", "https://example.com/" + strings.Repeat("é", 1014), true},
		{"over UTF-8 byte limit", "https://example.com/" + strings.Repeat("é", 1014) + "x", false},
		{"query key", "https://example.com/?q", false},
		{"query token", "https://example.com/?token=secret", false},
		{"query separator as content", "https://example.com/??", false},
		{"query whitespace", "https://example.com/?	", false},
		{"path control", "https://example.com/\n?", false},
		{"authority control", "https://exam	ple.com/?", false},
		{"credentials", "https://user:password@example.com/?", false},
		{"empty credentials", "https://@example.com/?", false},
		{"NUL", "https://example.com/\x00?", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := json.Marshal(StructuredDeliverable{
				SchemaVersion: 1, Kind: "research_report", Title: "Report",
				Sections: []DeliverableSection{{Heading: "Finding", Body: "Evidence"}},
				Sources:  []DeliverableSource{{Label: "Public reference", URL: tc.source}},
			})
			if err != nil {
				t.Fatal(err)
			}
			out, err := ParseDeliverable(raw)
			if !tc.valid {
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("unsafe source accepted: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Go-valid source rejected: %v", err)
			}
			if out.Sources[0].URL != tc.source {
				t.Fatalf("source normalized: got %q, want %q", out.Sources[0].URL, tc.source)
			}
		})
	}
}
