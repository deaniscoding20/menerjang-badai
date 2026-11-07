package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/menerjangbadai/dashboard/internal/application"
	"github.com/menerjangbadai/dashboard/internal/domain"
)

func taskFailure(w http.ResponseWriter, err error) {
	if errors.Is(err, domain.ErrTaskConflict) {
		problem(w, 409, "version_conflict", "Task changed; reload before saving")
		return
	}
	if errors.Is(err, domain.ErrTaskExecutionDisabled) {
		problem(w, 409, "execution_disabled", "Agent execution is unverified and disabled")
		return
	}
	if errors.Is(err, domain.ErrDependencyCycle) {
		problem(w, 409, "dependency_cycle", "Dependency would create a cycle")
		return
	}
	failure(w, err)
}

// NewTaskHandler must run behind Server.tasksEndpoint. A direct call has no actor.
func NewTaskHandler(service application.Tasks) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor, ok := taskActor(r.Context())
		if !ok {
			failure(w, domain.ErrUnauthorized)
			return
		}
		id := r.PathValue("id")
		switch {
		case id == "" && r.Method == "GET":
			tasks, err := service.List(r.Context(), actor, r.URL.Query().Get("brand_id"))
			if err != nil {
				taskFailure(w, err)
				return
			}
			send(w, 200, map[string]any{"items": tasks})
		case id == "" && r.Method == "POST":
			var in struct {
				BrandID      string     `json:"brand_id"`
				Title        string     `json:"title"`
				Objective    string     `json:"objective"`
				Instructions string     `json:"instructions"`
				Priority     string     `json:"priority"`
				AgentID      *string    `json:"agent_id"`
				DeadlineAt   *time.Time `json:"deadline_at"`
			}
			if !decode(w, r, &in) {
				return
			}
			agentID := ""
			if in.AgentID != nil {
				agentID = *in.AgentID
			}
			task, err := service.Create(r.Context(), actor, domain.Task{BrandID: in.BrandID, Title: in.Title, Objective: in.Objective, Instructions: in.Instructions, Priority: in.Priority, AssignedAgentID: agentID, DeadlineAt: in.DeadlineAt})
			if err != nil {
				taskFailure(w, err)
				return
			}
			w.Header().Set("Location", "/api/tasks/"+task.ID)
			send(w, 201, task)
		case id != "" && r.PathValue("deliverableID") == "" && strings.HasSuffix(r.URL.Path, "/revision-requests"):
			taskRevisionRequestsEndpoint(w, r, service, actor, id)
		case id != "" && (strings.HasSuffix(r.URL.Path, "/deliverables") || r.PathValue("deliverableID") != ""):
			taskDeliverablesEndpoint(w, r, service, actor, id)
		case id != "" && strings.HasSuffix(r.URL.Path, "/dependencies"):
			switch r.Method {
			case "GET":
				snapshot, err := service.Dependencies(r.Context(), actor, id)
				if err != nil {
					taskFailure(w, err)
					return
				}
				send(w, 200, snapshot)
			case "POST", "DELETE":
				var in domain.TaskDependencyChange
				if !decode(w, r, &in) {
					return
				}
				task, err := service.ChangeDependency(r.Context(), actor, id, in, r.Method == "POST")
				if err != nil {
					taskFailure(w, err)
					return
				}
				send(w, 200, task)
			default:
				w.Header().Set("Allow", "GET, POST, DELETE")
				problem(w, 405, "method_not_allowed", "Method not allowed")
			}
		case id != "" && strings.HasSuffix(r.URL.Path, "/comments"):
			switch r.Method {
			case "GET":
				comments, err := service.Comments(r.Context(), actor, id)
				if err != nil {
					taskFailure(w, err)
					return
				}
				send(w, 200, map[string]any{"items": comments})
			case "POST":
				var in struct {
					Body string `json:"body"`
					Kind string `json:"kind"`
				}
				if !decode(w, r, &in) {
					return
				}
				comment, err := service.AddComment(r.Context(), actor, id, domain.TaskComment{Body: in.Body, Kind: in.Kind})
				if err != nil {
					taskFailure(w, err)
					return
				}
				send(w, 201, comment)
			default:
				problem(w, 405, "method_not_allowed", "Method not allowed")
			}
		case id != "" && strings.HasSuffix(r.URL.Path, "/events") && r.Method == "GET":
			events, err := service.Events(r.Context(), actor, id)
			if err != nil {
				taskFailure(w, err)
				return
			}
			send(w, 200, map[string]any{"items": events})
		case id != "" && strings.HasSuffix(r.URL.Path, "/transition") && r.Method == "POST":
			var in domain.TaskTransition
			if !decode(w, r, &in) {
				return
			}
			task, err := service.Transition(r.Context(), actor, id, in)
			if err != nil {
				taskFailure(w, err)
				return
			}
			send(w, 200, task)
		case id != "" && !strings.HasSuffix(r.URL.Path, "/events") && !strings.HasSuffix(r.URL.Path, "/transition") && r.Method == "GET":
			task, err := service.Get(r.Context(), actor, id)
			if err != nil {
				taskFailure(w, err)
				return
			}
			send(w, 200, task)
		case id != "" && !strings.HasSuffix(r.URL.Path, "/events") && !strings.HasSuffix(r.URL.Path, "/transition") && r.Method == "PATCH":
			var in struct {
				Title        *string         `json:"title"`
				Objective    *string         `json:"objective"`
				Instructions *string         `json:"instructions"`
				Priority     *string         `json:"priority"`
				AgentID      json.RawMessage `json:"agent_id"`
				Version      int64           `json:"version"`
				DeadlineAt   json.RawMessage `json:"deadline_at"`
			}
			if !decode(w, r, &in) {
				return
			}
			current, err := service.Get(r.Context(), actor, id)
			if err != nil {
				taskFailure(w, err)
				return
			}
			if in.Version < 1 || in.Version != current.Version {
				taskFailure(w, domain.ErrTaskConflict)
				return
			}
			if in.Title != nil {
				current.Title = *in.Title
			}
			if in.Objective != nil {
				current.Objective = *in.Objective
			}
			if in.Instructions != nil {
				current.Instructions = *in.Instructions
			}
			if in.Priority != nil {
				current.Priority = *in.Priority
			}
			if in.AgentID != nil {
				if string(in.AgentID) == "null" {
					current.AssignedAgentID = ""
				} else if json.Unmarshal(in.AgentID, &current.AssignedAgentID) != nil {
					failure(w, domain.ErrInvalid)
					return
				}
			}
			if in.DeadlineAt != nil {
				if string(in.DeadlineAt) == "null" {
					current.DeadlineAt = nil
				} else {
					var deadline time.Time
					if json.Unmarshal(in.DeadlineAt, &deadline) != nil {
						failure(w, domain.ErrInvalid)
						return
					}
					current.DeadlineAt = &deadline
				}
			}
			task, err := service.Update(r.Context(), actor, current)
			if err != nil {
				taskFailure(w, err)
				return
			}
			send(w, 200, task)
		default:
			problem(w, 405, "method_not_allowed", "Method not allowed")
		}
	})
}
