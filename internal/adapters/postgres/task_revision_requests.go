package postgres

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/menerjangbadai/dashboard/internal/application"
	"github.com/menerjangbadai/dashboard/internal/domain"
)

var _ application.TaskRevisionRequestRepository = (*Store)(nil)

const revisionRequestColumns = `r.id,r.task_id,r.deliverable_id,r.deliverable_revision,r.author_id,r.body,r.created_at`

func scanRevisionRequest(row taskScanner) (domain.TaskRevisionRequest, error) {
	var out domain.TaskRevisionRequest
	err := row.Scan(&out.ID, &out.TaskID, &out.DeliverableID, &out.DeliverableRevision, &out.AuthorID, &out.Body, &out.CreatedAt)
	return out, safeErr(err)
}

func (s *Store) CreateTaskRevisionRequest(ctx context.Context, a domain.Actor, id string, in domain.RevisionRequestSubmission) (domain.TaskRevisionRequest, error) {
	validated, err := domain.NewRevisionRequestSubmission(in)
	if err != nil {
		return domain.TaskRevisionRequest{}, err
	}
	if _, err = uuid.Parse(id); err != nil {
		return domain.TaskRevisionRequest{}, domain.ErrNotFound
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return domain.TaskRevisionRequest{}, err
	}
	defer tx.Rollback(ctx)
	// Recheck and lock current membership, brand grant, write role and task state.
	task, err := lockTask(ctx, tx, a, id)
	if err != nil {
		return domain.TaskRevisionRequest{}, err
	}
	if task.Version != validated.Version {
		return domain.TaskRevisionRequest{}, domain.ErrTaskConflict
	}
	if !task.Editable() {
		return domain.TaskRevisionRequest{}, domain.ErrInvalid
	}
	// The target revision is derived from the immutable row, never caller input.
	// INSERT ... SELECT returns no row for a missing, foreign-parent or foreign-tenant target.
	saved, err := scanRevisionRequest(tx.QueryRow(ctx, `INSERT INTO task_revision_requests AS r
 (org_id,id,task_id,deliverable_id,deliverable_revision,author_id,body)
 SELECT $1,$2,$3,v.id,v.revision,$4,$5 FROM task_deliverable_versions v
 WHERE v.org_id=$1 AND v.task_id=$3 AND v.id=$6
 RETURNING `+revisionRequestColumns, a.OrgID, uuid.NewString(), task.ID, a.UserID, validated.Body, validated.DeliverableID))
	if err != nil {
		return domain.TaskRevisionRequest{}, err
	}
	updated, err := saveTask(ctx, tx, a, task)
	if err != nil {
		return domain.TaskRevisionRequest{}, err
	}
	// Feedback records a note ID, not body content, and leaves lifecycle state unchanged.
	if err = recordTaskEvent(ctx, tx, a, updated, "revision_requested", task.Status, saved.ID); err != nil {
		return domain.TaskRevisionRequest{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.TaskRevisionRequest{}, err
	}
	return saved, nil
}

func (s *Store) ListTaskRevisionRequests(ctx context.Context, a domain.Actor, id string) (domain.TaskRevisionRequests, error) {
	if _, err := uuid.Parse(id); err != nil {
		return domain.TaskRevisionRequests{}, domain.ErrNotFound
	}
	var out domain.TaskRevisionRequests
	var raw []byte
	// One SQL statement makes the visible task version and latest notes a single snapshot.
	err := s.Pool.QueryRow(ctx, `SELECT t.version,COALESCE((SELECT jsonb_agg(jsonb_build_object(
 'id',r.id,'task_id',r.task_id,'deliverable_id',r.deliverable_id,'deliverable_revision',r.deliverable_revision,
 'author_id',r.author_id,'body',r.body,'created_at',r.created_at
 ) ORDER BY r.created_at DESC,r.id DESC) FROM (
 SELECT * FROM task_revision_requests n WHERE n.org_id=t.org_id AND n.task_id=t.id
 ORDER BY n.created_at DESC,n.id DESC LIMIT 100
 ) r),'[]'::jsonb)`+taskAccess+` AND t.id=$3`, a.OrgID, a.UserID, id).Scan(&out.Version, &raw)
	if err != nil {
		return domain.TaskRevisionRequests{}, safeErr(err)
	}
	if err = json.Unmarshal(raw, &out.Items); err != nil {
		return domain.TaskRevisionRequests{}, err
	}
	return out, nil
}
