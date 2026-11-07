package httpapi

import (
	"context"
	"encoding/json"
	"github.com/menerjangbadai/dashboard/internal/application"
	"github.com/menerjangbadai/dashboard/internal/domain"
	"net/http"
	"strings"
	"testing"
	"time"
)

// Transport fixture only: these assertions do not certify real DB isolation.
type revisionHTTPRepo struct {
	taskHTTPRepo
	role      string
	hidden    bool
	saved     int
	listed    int
	targetErr error
	items     []domain.TaskRevisionRequest
}

func (r *revisionHTTPRepo) Membership(_ context.Context, a domain.Actor) (domain.Actor, error) {
	a.Role = r.role
	return a, nil
}
func (r *revisionHTTPRepo) GetTask(context.Context, domain.Actor, string) (domain.Task, error) {
	if r.hidden {
		return domain.Task{}, domain.ErrNotFound
	}
	return r.task, nil
}
func (r *revisionHTTPRepo) ListTaskRevisionRequests(context.Context, domain.Actor, string) (domain.TaskRevisionRequests, error) {
	r.listed++
	return domain.TaskRevisionRequests{Version: r.task.Version, Items: r.items}, nil
}
func (r *revisionHTTPRepo) CreateTaskRevisionRequest(_ context.Context, a domain.Actor, id string, in domain.RevisionRequestSubmission) (domain.TaskRevisionRequest, error) {
	if r.targetErr != nil {
		return domain.TaskRevisionRequest{}, r.targetErr
	}
	item := domain.TaskRevisionRequest{ID: httpDeliverableID, TaskID: id, DeliverableID: in.DeliverableID, DeliverableRevision: 2, AuthorID: a.UserID, Body: in.Body, CreatedAt: time.Now()}
	r.saved++
	r.task.Version++
	r.items = append(r.items, item)
	return item, nil
}
func revisionFixture() (http.Handler, *revisionHTTPRepo, string, string) {
	repo := &revisionHTTPRepo{taskHTTPRepo: taskHTTPRepo{task: domain.Task{ID: httpDependencySource, Status: domain.TaskReady, Version: 3}}, role: "editor", items: []domain.TaskRevisionRequest{}}
	h := New(Config{PublicURL: "https://app.test", CommitSHA: strings.Repeat("a", 40)}, application.Brands{}, &authMemory{}, provider{}, func(context.Context) error { return nil }, NewTaskHandler(application.Tasks{Repo: repo}))
	return h, repo, "/api/tasks/" + httpDependencySource + "/revision-requests", application.CSRF("opaque-test-session")
}
func TestTaskRevisionRequestHTTPListCreateAndStale(t *testing.T) {
	h, repo, path, csrf := revisionFixture()
	if res := call(h, "GET", path, "", "", "", false); res.Code != 401 || repo.listed != 0 {
		t.Fatal("session", res.Code)
	}
	if res := call(h, "GET", path, "", "", "", true); res.Code != 200 || !strings.Contains(res.Body.String(), `"version":3`) || !strings.Contains(res.Body.String(), `"items":[]`) {
		t.Fatal("snapshot", res.Code, res.Body.String())
	}
	body := `{"deliverable_id":"` + httpDeliverableID + `","body":"  Human request.  ","version":3}`
	for _, tc := range []struct{ origin, token string }{{"https://evil.test", csrf}, {"", csrf}, {"https://app.test", ""}, {"https://app.test", "bad"}} {
		if res := call(h, "POST", path, body, tc.origin, tc.token, true); res.Code != 403 || repo.saved != 0 {
			t.Fatal("csrf", res.Code)
		}
	}
	res := call(h, "POST", path, body, "https://app.test", csrf, true)
	var got domain.TaskRevisionRequest
	if res.Code != 201 || json.Unmarshal(res.Body.Bytes(), &got) != nil || got.Body != "Human request." || got.TaskID != httpDependencySource || got.DeliverableID != httpDeliverableID || got.DeliverableRevision != 2 || got.AuthorID != "user" || repo.saved != 1 || repo.task.Version != 4 || repo.task.Status != domain.TaskReady {
		t.Fatal("create", res.Code, res.Body.String(), repo)
	}
	if res = call(h, "POST", path, body, "https://app.test", csrf, true); res.Code != 409 || !strings.Contains(res.Body.String(), "version_conflict") || repo.saved != 1 {
		t.Fatal("stale", res.Code)
	}
	repo.role = "viewer"
	if res = call(h, "GET", path, "", "", "", true); res.Code != 200 || !strings.Contains(res.Body.String(), `"version":4`) {
		t.Fatal("readonly list", res.Code)
	}
}
func TestTaskRevisionRequestHTTPAuthorityAndInput(t *testing.T) {
	cases := []struct {
		name, body, role string
		status           domain.TaskStatus
		want             int
		hidden           bool
		targetErr        error
	}{
		{name: "forged author", body: `{"deliverable_id":"` + httpDeliverableID + `","body":"note","version":3,"author_id":"forged"}`, want: 400},
		{name: "forged org", body: `{"deliverable_id":"` + httpDeliverableID + `","body":"note","version":3,"org_id":"forged"}`, want: 400},
		{name: "trailing JSON", body: `{"deliverable_id":"` + httpDeliverableID + `","body":"note","version":3} {}`, want: 400},
		{name: "missing version", body: `{"deliverable_id":"` + httpDeliverableID + `","body":"note"}`, want: 409},
		{name: "invalid target", body: `{"deliverable_id":"bad","body":"note","version":3}`, want: 404},
		{name: "blank", body: `{"deliverable_id":"` + httpDeliverableID + `","body":"\u0085","version":3}`, want: 422},
		{name: "NUL", body: `{"deliverable_id":"` + httpDeliverableID + `","body":"note\u0000","version":3}`, want: 422},
		{name: "invalid UTF8", body: `{"deliverable_id":"` + httpDeliverableID + `","body":"raw` + string([]byte{0xff}) + `","version":3}`, want: 400},
		{name: "4001 emoji", body: `{"deliverable_id":"` + httpDeliverableID + `","body":"` + strings.Repeat("🌀", 4001) + `","version":3}`, want: 422},
		{name: "oversize", body: strings.Repeat("x", 65537), want: 400},
		{name: "hidden parent", body: `{"deliverable_id":"` + httpDeliverableID + `","body":"note","version":3}`, hidden: true, want: 404},
		{name: "wrong target parent", body: `{"deliverable_id":"` + httpDeliverableID + `","body":"note","version":3}`, targetErr: domain.ErrNotFound, want: 404},
	}
	for _, role := range []string{"viewer", "approver", "analyst"} {
		cases = append(cases, struct {
			name, body, role string
			status           domain.TaskStatus
			want             int
			hidden           bool
			targetErr        error
		}{name: role, role: role, body: `{"deliverable_id":"` + httpDeliverableID + `","body":"note","version":3}`, want: 403})
	}
	for _, status := range []domain.TaskStatus{domain.TaskQueued, domain.TaskRunning, domain.TaskWaitingInput, domain.TaskWaitingApproval, domain.TaskCompleted, domain.TaskFailed, domain.TaskCancelled} {
		cases = append(cases, struct {
			name, body, role string
			status           domain.TaskStatus
			want             int
			hidden           bool
			targetErr        error
		}{name: string(status), status: status, body: `{"deliverable_id":"` + httpDeliverableID + `","body":"note","version":3}`, want: 422})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, r, path, csrf := revisionFixture()
			if tc.role != "" {
				r.role = tc.role
			}
			if tc.status != "" {
				r.task.Status = tc.status
			}
			r.hidden = tc.hidden
			r.targetErr = tc.targetErr
			res := call(h, "POST", path, tc.body, "https://app.test", csrf, true)
			if res.Code != tc.want || r.saved != 0 || r.task.Version != 3 {
				t.Fatal(res.Code, tc.want, res.Body.String(), r.saved, r.task.Version)
			}
		})
	}
	for _, method := range []string{"PATCH", "DELETE", "PUT"} {
		h, r, path, csrf := revisionFixture()
		if res := call(h, method, path, `{}`, "https://app.test", csrf, true); res.Code != 405 || r.saved != 0 {
			t.Fatal(method, res.Code)
		}
	}
}
func TestTaskRevisionRequestHTTPUnicodeEscapeOverhead(t *testing.T) {
	for _, body := range []string{strings.Repeat("🌀", 4000), strings.Repeat(`\ud83c\udf00`, 4000), "\ufffd🌀", "\ufeff"} {
		h, r, path, csrf := revisionFixture()
		// Escape sequences in the second case must reach the decoder unchanged.
		raw := `{"deliverable_id":"` + httpDeliverableID + `","body":"` + body + `","version":3}`
		res := call(h, "POST", path, raw, "https://app.test", csrf, true)
		if res.Code != 201 || r.saved != 1 {
			t.Fatal("valid Unicode body", len(raw), res.Code, res.Body.String())
		}
	}
}
