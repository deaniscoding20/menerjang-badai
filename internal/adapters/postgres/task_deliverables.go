package postgres

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/menerjangbadai/dashboard/internal/application"
	"github.com/menerjangbadai/dashboard/internal/domain"
)

var _ application.TaskDeliverableRepository = (*Store)(nil)

const deliverableColumns = `v.id,v.task_id,v.revision,v.author_id,v.origin,v.validation_status,COALESCE(v.structured->>'title',''),COALESCE(v.structured->>'kind',''),v.created_at,v.raw_output,v.structured`

func scanDeliverable(row taskScanner) (domain.DeliverableVersion, error) {
	var out domain.DeliverableVersion
	var structured []byte
	err := row.Scan(&out.ID, &out.TaskID, &out.Revision, &out.AuthorID, &out.Origin, &out.ValidationStatus, &out.Title, &out.Kind, &out.CreatedAt, &out.RawOutput, &structured)
	if err != nil {
		return out, safeErr(err)
	}
	if structured != nil {
		if err = json.Unmarshal(structured, &out.Structured); err != nil {
			return domain.DeliverableVersion{}, err
		}
	}
	return out, nil
}
func (s *Store) SubmitTaskDeliverable(ctx context.Context, a domain.Actor, id string, in domain.DeliverableSubmission) (domain.DeliverableVersion, error) {
	validated, err := domain.NewDeliverableSubmission(in)
	if err != nil {
		return domain.DeliverableVersion{}, err
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return domain.DeliverableVersion{}, err
	}
	defer tx.Rollback(ctx)
	task, err := lockTask(ctx, tx, a, id)
	if err != nil {
		return domain.DeliverableVersion{}, err
	}
	if task.Version != in.Version {
		return domain.DeliverableVersion{}, domain.ErrTaskConflict
	}
	if !task.Editable() {
		return domain.DeliverableVersion{}, domain.ErrInvalid
	}
	var structured any
	if validated.Structured != nil {
		encoded, e := json.Marshal(validated.Structured)
		if e != nil {
			return domain.DeliverableVersion{}, e
		}
		structured = string(encoded)
	}
	// lockTask serializes revision allocation with all imports on this task.
	saved, err := scanDeliverable(tx.QueryRow(ctx, `INSERT INTO task_deliverable_versions AS v(org_id,id,task_id,revision,author_id,origin,raw_output,validation_status,structured)
 VALUES($1,$2,$3,(SELECT COALESCE(max(revision),0)+1 FROM task_deliverable_versions WHERE org_id=$1 AND task_id=$3),$4,'human_import',$5,$6,$7::jsonb)
 RETURNING `+deliverableColumns, a.OrgID, uuid.NewString(), task.ID, a.UserID, validated.RawOutput, validated.ValidationStatus, structured))
	if err != nil {
		return domain.DeliverableVersion{}, err
	}
	updated, err := saveTask(ctx, tx, a, task)
	if err != nil {
		return domain.DeliverableVersion{}, err
	}
	if err = recordTaskEvent(ctx, tx, a, updated, "deliverable_submitted", task.Status, saved.ID); err != nil {
		return domain.DeliverableVersion{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.DeliverableVersion{}, err
	}
	return saved, nil
}
func (s *Store) GetTaskDeliverable(ctx context.Context, a domain.Actor, id string, versionID string) (domain.DeliverableVersion, error) {
	return scanDeliverable(s.Pool.QueryRow(ctx, `SELECT `+deliverableColumns+` FROM task_deliverable_versions v
 WHERE v.org_id=$1 AND v.task_id=$3 AND v.id=$4 AND EXISTS(SELECT 1`+taskAccess+` AND t.id=$3)`, a.OrgID, a.UserID, id, versionID))
}
func (s *Store) ListTaskDeliverables(ctx context.Context, a domain.Actor, id string) (domain.TaskDeliverables, error) {
	var out domain.TaskDeliverables
	var raw []byte
	// Revision metadata only; raw/structured bodies are fetched individually.
	err := s.Pool.QueryRow(ctx, `SELECT t.version,COALESCE((SELECT jsonb_agg(jsonb_build_object(
 'id',v.id,'task_id',v.task_id,'revision',v.revision,'author_id',v.author_id,'origin',v.origin,
 'validation_status',v.validation_status,'title',COALESCE(v.structured->>'title',''),'kind',COALESCE(v.structured->>'kind',''),'created_at',v.created_at
 ) ORDER BY v.revision DESC) FROM (SELECT * FROM task_deliverable_versions d WHERE d.org_id=t.org_id AND d.task_id=t.id ORDER BY d.revision DESC LIMIT 100) v),'[]'::jsonb)`+taskAccess+` AND t.id=$3`, a.OrgID, a.UserID, id).Scan(&out.Version, &raw)
	if err != nil {
		return domain.TaskDeliverables{}, safeErr(err)
	}
	if err = json.Unmarshal(raw, &out.Items); err != nil {
		return domain.TaskDeliverables{}, err
	}
	return out, nil
}
