package httpapi

import (
	"github.com/menerjangbadai/dashboard/internal/application"
	"github.com/menerjangbadai/dashboard/internal/domain"
	"net/http"
)

func taskDeliverablesEndpoint(w http.ResponseWriter, r *http.Request, service application.Tasks, actor domain.Actor, id string) {
	versionID := r.PathValue("deliverableID")
	switch {
	case r.Method == "GET" && versionID != "":
		item, err := service.Deliverable(r.Context(), actor, id, versionID)
		if err != nil {
			taskFailure(w, err)
			return
		}
		send(w, 200, item)
	case r.Method == "GET":
		snapshot, err := service.Deliverables(r.Context(), actor, id)
		if err != nil {
			taskFailure(w, err)
			return
		}
		send(w, 200, snapshot)
	case r.Method == "POST" && versionID == "":
		var in domain.DeliverableSubmission
		// Permit JSON escaping overhead without widening other API request limits.
		if !decodeLimit(w, r, &in, 1024*1024) {
			return
		}
		item, err := service.SubmitDeliverable(r.Context(), actor, id, in)
		if err != nil {
			taskFailure(w, err)
			return
		}
		w.Header().Set("Location", "/api/tasks/"+item.TaskID+"/deliverables/"+item.ID)
		send(w, 201, item)
	default:
		w.Header().Set("Allow", "GET")
		if versionID == "" {
			w.Header().Set("Allow", "GET, POST")
		}
		problem(w, 405, "method_not_allowed", "Method not allowed")
	}
}
