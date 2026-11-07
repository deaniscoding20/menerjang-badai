package domain

import (
	"bytes"
	"encoding/json"
	"io"
	"net/url"
	"strings"
	"unicode/utf8"
)

// StructuredDeliverable v1 is an envelope for reviewed, structured content.
// Parsing does not approve its facts, publish it, or mark a task completed.
type StructuredDeliverable struct {
	SchemaVersion int                  `json:"schema_version"`
	Kind          string               `json:"kind"`
	Title         string               `json:"title"`
	Sections      []DeliverableSection `json:"sections"`
	Sources       []DeliverableSource  `json:"sources,omitempty"`
}
type DeliverableSection struct {
	Heading string `json:"heading"`
	Body    string `json:"body"`
}
type DeliverableSource struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

func ParseDeliverable(raw json.RawMessage) (StructuredDeliverable, error) {
	var out StructuredDeliverable
	if len(raw) == 0 || len(raw) > 128*1024 {
		return out, ErrInvalid
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&out) != nil {
		return out, ErrInvalid
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return out, ErrInvalid
	}
	out.Title = strings.TrimSpace(out.Title)
	if out.SchemaVersion != 1 || out.Title == "" || utf8.RuneCountInString(out.Title) > 200 || len(out.Sections) < 1 || len(out.Sections) > 40 || len(out.Sources) > 40 {
		return out, ErrInvalid
	}
	switch out.Kind {
	case "research_report", "marketing_strategy", "creative_brief", "caption_package", "design_specification", "campaign_plan", "performance_report":
	default:
		return out, ErrInvalid
	}
	for _, section := range out.Sections {
		if strings.TrimSpace(section.Heading) == "" || utf8.RuneCountInString(section.Heading) > 160 || strings.TrimSpace(section.Body) == "" || utf8.RuneCountInString(section.Body) > 12000 {
			return out, ErrInvalid
		}
	}
	for _, source := range out.Sources {
		if strings.TrimSpace(source.Label) == "" || utf8.RuneCountInString(source.Label) > 200 || len(source.URL) > 2048 {
			return out, ErrInvalid
		}
		u, err := url.Parse(source.URL)
		// Store public references only, never temporary signed URLs or embedded credentials.
		if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.RawQuery != "" {
			return out, ErrInvalid
		}
	}
	return out, nil
}
