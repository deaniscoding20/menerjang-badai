package httpapi

import (
	"strings"
	"testing"
)

func TestTaskRoutesRequireAuthenticationWhenExecutionUnavailable(t *testing.T) {
	handler, _ := setup(t)
	for _, item := range []struct{ method, path, body string }{
		{"GET", "/api/tasks", ""},
		{"POST", "/api/tasks", `{"brand_id":"d0e0ddbb-23f8-4fe7-9be7-e53dd92e9860","title":"Brief"}`},
		{"GET", "/api/tasks/d0e0ddbb-23f8-4fe7-9be7-e53dd92e9860", ""},
		{"PATCH", "/api/tasks/d0e0ddbb-23f8-4fe7-9be7-e53dd92e9860", `{"title":"Edit"}`},
		{"GET", "/api/tasks/d0e0ddbb-23f8-4fe7-9be7-e53dd92e9860/events", ""},
		{"POST", "/api/tasks/d0e0ddbb-23f8-4fe7-9be7-e53dd92e9860/transition", `{"status":"running"}`},
	} {
		response := call(handler, item.method, item.path, item.body, "", "", false)
		if response.Code != 401 {
			t.Fatalf("%s %s: want401 got%d %s", item.method, item.path, response.Code, response.Body.String())
		}
	}
}

func TestTaskMutationRejectsCSRFBeforeUnconfiguredRepository(t *testing.T) {
	handler, _ := setup(t)
	response := call(handler, "POST", "/api/tasks", `{"title":"Brief"}`, "https://app.test", "", true)
	if response.Code != 403 || !strings.Contains(response.Body.String(), "csrf_rejected") {
		t.Fatalf("%d %s", response.Code, response.Body.String())
	}
}
