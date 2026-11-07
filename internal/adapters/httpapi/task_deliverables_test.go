package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/menerjangbadai/dashboard/internal/application"
	"github.com/menerjangbadai/dashboard/internal/domain"
	"strings"
	"testing"
)

const httpDeliverableID = "618d5470-f095-4e3a-8a04-3bfe518b7c25"

type deliverableHTTPRepo struct {
	taskHTTPRepo
	saved domain.DeliverableVersion
}

func (r *deliverableHTTPRepo) ListTaskDeliverables(context.Context, domain.Actor, string) (domain.TaskDeliverables, error) {
	return domain.TaskDeliverables{Version: r.task.Version, Items: []domain.DeliverableSummary{}}, nil
}
func (r *deliverableHTTPRepo) GetTaskDeliverable(context.Context, domain.Actor, string, string) (domain.DeliverableVersion, error) {
	return r.saved, nil
}
func (r *deliverableHTTPRepo) SubmitTaskDeliverable(_ context.Context, a domain.Actor, id string, in domain.DeliverableSubmission) (domain.DeliverableVersion, error) {
	parsed, err := domain.NewDeliverableSubmission(in)
	if err != nil {
		return domain.DeliverableVersion{}, err
	}
	r.saved = domain.DeliverableVersion{DeliverableSummary: domain.DeliverableSummary{ID: httpDeliverableID, TaskID: id, Revision: 1, AuthorID: a.UserID, Origin: "human_import", ValidationStatus: parsed.ValidationStatus}, RawOutput: parsed.RawOutput, Structured: parsed.Structured}
	r.task.Version++
	return r.saved, nil
}
func TestTaskDeliverableHTTPRejectsInvalidUTF8WithoutWrite(t *testing.T) {
	repo := &deliverableHTTPRepo{taskHTTPRepo: taskHTTPRepo{task: domain.Task{ID: httpDependencySource, BrandID: "d0e0ddbb-23f8-4fe7-9be7-e53dd92e9860", Status: domain.TaskReady, Version: 3}}}
	handler := New(Config{PublicURL: "https://app.test", CommitSHA: strings.Repeat("a", 40)}, application.Brands{}, &authMemory{}, provider{}, func(context.Context) error { return nil }, NewTaskHandler(application.Tasks{Repo: repo}))
	raw := "raw\xffoutput"
	if _, err := domain.NewDeliverableSubmission(domain.DeliverableSubmission{RawOutput: raw, Version: 3}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatal("domain must reject the original invalid UTF-8")
	}
	// Build the JSON literally: json.Marshal would replace the invalid byte first.
	body := `{"raw_output":"` + raw + `","version":3}`
	res := call(handler, "POST", "/api/tasks/"+httpDependencySource+"/deliverables", body, "https://app.test", application.CSRF("opaque-test-session"), true)
	if res.Code != 400 || repo.saved.ID != "" || repo.task.Version != 3 {
		t.Fatalf("invalid UTF-8 import: status=%d want=400 saved=%t version=%d want=3 raw_altered=%t", res.Code, repo.saved.ID != "", repo.task.Version, repo.saved.ID != "" && repo.saved.RawOutput != raw)
	}
	if !strings.Contains(res.Body.String(), `"code":"invalid_json"`) || strings.Contains(res.Body.String(), "raw") {
		t.Fatal("expected safe invalid_json response")
	}
}

func TestTaskDeliverableHTTPRawImportAndBoundaries(t *testing.T) {
	repo := &deliverableHTTPRepo{taskHTTPRepo: taskHTTPRepo{task: domain.Task{ID: httpDependencySource, BrandID: "d0e0ddbb-23f8-4fe7-9be7-e53dd92e9860", Status: domain.TaskReady, Version: 3}}}
	handler := New(Config{PublicURL: "https://app.test", CommitSHA: strings.Repeat("a", 40)}, application.Brands{}, &authMemory{}, provider{}, func(context.Context) error { return nil }, NewTaskHandler(application.Tasks{Repo: repo}))
	path := "/api/tasks/" + httpDependencySource + "/deliverables"
	csrf := application.CSRF("opaque-test-session")
	raw := strings.Repeat("raw output\n", 4000) + "literal \ufffd emoji 🌀"
	body, _ := json.Marshal(domain.DeliverableSubmission{RawOutput: raw, Version: 3})
	if res := call(handler, "GET", path, "", "", "", false); res.Code != 401 {
		t.Fatal("session", res.Code)
	}
	if res := call(handler, "GET", path, "", "", "", true); res.Code != 200 || !strings.Contains(res.Body.String(), `"version":3`) {
		t.Fatal("list", res.Code, res.Body.String())
	}
	if res := call(handler, "POST", path, string(body), "https://app.test", "bad", true); res.Code != 403 || repo.task.Version != 3 {
		t.Fatal("csrf", res.Code)
	}
	if res := call(handler, "POST", path, `{"raw_output":"raw","version":3,"author_id":"forged"}`, "https://app.test", csrf, true); res.Code != 400 {
		t.Fatal("forged author", res.Code)
	}
	if res := call(handler, "POST", path, string(body), "https://app.test", csrf, true); res.Code != 201 || repo.saved.RawOutput != raw || repo.saved.ValidationStatus != "invalid" || repo.saved.Structured != nil {
		t.Fatal("import", res.Code, res.Body.String())
	}
	if repo.task.Status != domain.TaskReady || repo.task.Version != 4 {
		t.Fatal("import lifecycle", repo.task)
	}
	if res := call(handler, "POST", path, string(body), "https://app.test", csrf, true); res.Code != 409 {
		t.Fatal("stale", res.Code)
	}
	if res := call(handler, "GET", path+"/"+httpDeliverableID, "", "", "", true); res.Code != 200 || !strings.Contains(res.Body.String(), "human_import") {
		t.Fatal("detail", res.Code)
	}
	tooLarge, _ := json.Marshal(domain.DeliverableSubmission{RawOutput: strings.Repeat("x", 128*1024+1), Version: 4})
	if res := call(handler, "POST", path, string(tooLarge), "https://app.test", csrf, true); res.Code != 422 {
		t.Fatal("raw byte cap", res.Code)
	}
	if res := call(handler, "POST", path, strings.Repeat("x", 1024*1024+1), "https://app.test", csrf, true); res.Code != 400 {
		t.Fatal("request body cap", res.Code)
	}
}
