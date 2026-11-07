package httpapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/menerjangbadai/dashboard/internal/domain"
)

type taskActorKey struct{}

// The task handler receives authority only after session and CSRF checks.
// Request fields cannot replace this private server-owned context value.
func (s *Server) tasksEndpoint(w http.ResponseWriter, r *http.Request) {
	dependencyDelete := r.Method == "DELETE" && strings.HasSuffix(r.URL.Path, "/dependencies")
	if r.Method != "GET" && r.Method != "POST" && r.Method != "PATCH" && !dependencyDelete {
		w.Header().Set("Allow", "GET, POST, PATCH")
		problem(w, 405, "method_not_allowed", "Method not allowed")
		return
	}
	session, _, ok := s.current(w, r, r.Method != "GET")
	if !ok {
		return
	}
	if s.taskAPI == nil {
		problem(w, 503, "tasks_unavailable", "Task persistence is not configured")
		return
	}
	s.taskAPI.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), taskActorKey{}, session.Actor)))
}

func taskActor(ctx context.Context) (domain.Actor, bool) {
	a, ok := ctx.Value(taskActorKey{}).(domain.Actor)
	return a, ok
}
