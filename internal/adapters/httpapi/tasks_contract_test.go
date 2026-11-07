package httpapi

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/menerjangbadai/dashboard/internal/application"
)

func TestTaskHandlerReceivesOnlyServerSessionAuthority(t *testing.T) {
	invoked := false
	transport := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor, ok := taskActor(r.Context())
		if !ok || actor.UserID != "user" || actor.OrgID != "org" || actor.Role != "owner" {
			t.Errorf("no server-owned identity: %+v", actor)
		}
		if r.PathValue("id") != "d0e0ddbb-23f8-4fe7-9be7-e53dd92e9860" {
			t.Error("lost path value")
		}
		invoked = true
		send(w, 200, map[string]string{"id": r.PathValue("id")})
	})
	handler := New(Config{PublicURL: "https://app.test", CommitSHA: strings.Repeat("a", 40)}, application.Brands{}, &authMemory{}, provider{}, func(context.Context) error { return nil }, transport)
	response := call(handler, "GET", "/api/tasks/d0e0ddbb-23f8-4fe7-9be7-e53dd92e9860?org_id=forged", "", "", "", true)
	if response.Code != 200 || !invoked {
		t.Fatalf("%d %s", response.Code, response.Body.String())
	}
}

func TestUnconfiguredTasksNeverReturnFixtureSuccess(t *testing.T) {
	handler, _ := setup(t)
	response := call(handler, "GET", "/api/tasks", "", "", "", true)
	if response.Code != 503 || !strings.Contains(response.Body.String(), "tasks_unavailable") {
		t.Fatalf("%d %s", response.Code, response.Body.String())
	}
}
