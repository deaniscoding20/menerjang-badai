package httpapi

import (
	"net/http"
	"strings"
	"testing"
)

func TestTaskRevisionRequestHTTPDeliverableRouteCollision(t *testing.T) {
	for _, tc := range []struct {
		method string
		status int
		code   string
	}{
		{http.MethodGet, http.StatusNotFound, "not_found"},
		{http.MethodPost, http.StatusMethodNotAllowed, "method_not_allowed"},
	} {
		t.Run(tc.method, func(t *testing.T) {
			h, repo, _, csrf := revisionFixture()
			path := "/api/tasks/" + httpDependencySource + "/deliverables/revision-requests"
			body := `{"deliverable_id":"` + httpDeliverableID + `","body":"note","version":3}`
			res := call(h, tc.method, path, body, "https://app.test", csrf, true)
			if res.Code != tc.status || !strings.Contains(res.Body.String(), `"code":"`+tc.code+`"`) {
				t.Errorf("deliverable route: got %d %s; want %d %s", res.Code, res.Body.String(), tc.status, tc.code)
			}
			if tc.method == http.MethodPost && res.Header().Get("Allow") != "GET" {
				t.Errorf("deliverable detail Allow = %q; want GET", res.Header().Get("Allow"))
			}
			if repo.listed != 0 || repo.saved != 0 || repo.task.Version != 3 || len(repo.items) != 0 {
				t.Errorf("revision side effects: listed=%d saved=%d version=%d items=%d; want 0 0 3 0", repo.listed, repo.saved, repo.task.Version, len(repo.items))
			}
		})
	}
}

func TestTaskRevisionRequestHTTPGenuineRouteStillWorks(t *testing.T) {
	h, repo, path, csrf := revisionFixture()
	if res := call(h, http.MethodGet, path, "", "", "", true); res.Code != http.StatusOK || repo.listed != 1 || repo.saved != 0 || repo.task.Version != 3 {
		t.Fatalf("revision list: got %d %s; listed=%d saved=%d version=%d", res.Code, res.Body.String(), repo.listed, repo.saved, repo.task.Version)
	}
	body := `{"deliverable_id":"` + httpDeliverableID + `","body":"note","version":3}`
	if res := call(h, http.MethodPost, path, body, "https://app.test", csrf, true); res.Code != http.StatusCreated || repo.listed != 1 || repo.saved != 1 || repo.task.Version != 4 || len(repo.items) != 1 {
		t.Fatalf("revision create: got %d %s; listed=%d saved=%d version=%d items=%d", res.Code, res.Body.String(), repo.listed, repo.saved, repo.task.Version, len(repo.items))
	}
}
