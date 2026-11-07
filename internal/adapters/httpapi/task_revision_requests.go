package httpapi

import (
	"github.com/menerjangbadai/dashboard/internal/application"
	"github.com/menerjangbadai/dashboard/internal/domain"
	"net/http"
)

func taskRevisionRequestsEndpoint(w http.ResponseWriter, r *http.Request, service application.Tasks, actor domain.Actor, id string) {
	switch r.Method {
	case "GET":
		snapshot, err := service.RevisionRequests(r.Context(), actor, id)
		if err != nil {
			taskFailure(w, err)
			return
		}
		send(w, 200, snapshot)
	case "POST":
		var in domain.RevisionRequestSubmission
		// 4000 supplementary-plane runes may occupy 48000 JSON escape bytes.
		// Keep the larger cap confined to this endpoint, not the generic decoder.
		if !decodeLimit(w, r, &in, 65536) {
			return
		}
		item, err := service.RequestRevision(r.Context(), actor, id, in)
		if err != nil {
			taskFailure(w, err)
			return
		}
		send(w, 201, item)
	default:
		w.Header().Set("Allow", "GET, POST")
		problem(w, 405, "method_not_allowed", "Method not allowed")
	}
}
