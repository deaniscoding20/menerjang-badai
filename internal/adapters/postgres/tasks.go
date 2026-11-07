package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/menerjangbadai/dashboard/internal/application"
	"github.com/menerjangbadai/dashboard/internal/domain"
)

var _ application.TaskRepository = (*Store)(nil)

const taskColumns = `t.id,t.org_id,t.brand_id,t.requester_id,t.title,t.objective,t.instructions,t.priority,t.status,t.agent_id,t.version,t.deadline_at,t.created_at,t.updated_at`
const taskAccess = ` FROM tasks t
 JOIN brands b ON b.org_id=t.org_id AND b.id=t.brand_id
 JOIN user_brand_access g ON g.org_id=t.org_id AND g.brand_id=t.brand_id AND g.user_id=$2 AND g.active
 JOIN organization_members m ON m.org_id=g.org_id AND m.user_id=g.user_id AND m.active
 JOIN users u ON u.id=m.user_id AND u.active
 JOIN organizations o ON o.id=m.org_id AND o.active
 JOIN user_roles r ON r.org_id=m.org_id AND r.user_id=m.user_id
 WHERE t.org_id=$1`

type taskScanner interface{ Scan(...any) error }

func scanTask(row taskScanner) (domain.Task, error) {
	var task domain.Task
	err := row.Scan(&task.ID, &task.OrgID, &task.BrandID, &task.RequesterID, &task.Title, &task.Objective, &task.Instructions, &task.Priority, &task.Status, &task.AssignedAgentID, &task.Version, &task.DeadlineAt, &task.CreatedAt, &task.UpdatedAt)
	return task, safeErr(err)
}
func (s *Store) ListTasks(ctx context.Context, a domain.Actor, brandID string) ([]domain.Task, error) {
	var filter any
	if brandID != "" {
		filter = brandID
	}
	rows, err := s.Pool.Query(ctx, "SELECT "+taskColumns+taskAccess+` AND ($3::uuid IS NULL OR t.brand_id=$3) ORDER BY t.updated_at DESC,t.id LIMIT 500`, a.OrgID, a.UserID, filter)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, task)
	}
	return out, rows.Err()
}
func (s *Store) GetTask(ctx context.Context, a domain.Actor, id string) (domain.Task, error) {
	return scanTask(s.Pool.QueryRow(ctx, "SELECT "+taskColumns+taskAccess+` AND t.id=$3`, a.OrgID, a.UserID, id))
}
func (s *Store) CreateTask(ctx context.Context, a domain.Actor, in domain.Task) (domain.Task, error) {
	var err error
	in, err = domain.NewTask(in)
	if err != nil {
		return domain.Task{}, err
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return domain.Task{}, err
	}
	defer tx.Rollback(ctx)
	// Lock current membership, role and grant during this transaction. Claimed caller role is never authority.
	var role, status string
	err = tx.QueryRow(ctx, `SELECT r.role_id,b.status FROM brands b
 JOIN user_brand_access g ON g.org_id=b.org_id AND g.brand_id=b.id AND g.user_id=$2 AND g.active
 JOIN organization_members m ON m.org_id=g.org_id AND m.user_id=g.user_id AND m.active
 JOIN user_roles r ON r.org_id=m.org_id AND r.user_id=m.user_id
 JOIN users u ON u.id=m.user_id AND u.active JOIN organizations o ON o.id=m.org_id AND o.active
 WHERE b.org_id=$1 AND b.id=$3 FOR SHARE OF b,g,m,r,u,o`, a.OrgID, a.UserID, in.BrandID).Scan(&role, &status)
	if err != nil {
		return domain.Task{}, safeErr(err)
	}
	if !taskWriteRole(role) {
		return domain.Task{}, domain.ErrForbidden
	}
	if status != "active" {
		return domain.Task{}, domain.ErrInvalid
	}
	a.Role = role
	in.ID = uuid.NewString()
	in.OrgID = a.OrgID
	in.RequesterID = a.UserID
	saved, err := scanTask(tx.QueryRow(ctx, `INSERT INTO tasks AS t(org_id,id,brand_id,requester_id,title,objective,instructions,priority,agent_id,deadline_at)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING `+taskColumns, a.OrgID, in.ID, in.BrandID, a.UserID, in.Title, in.Objective, in.Instructions, in.Priority, in.AssignedAgentID, in.DeadlineAt))
	if err != nil {
		return domain.Task{}, err
	}
	if err = recordTaskEvent(ctx, tx, a, saved, "created", "", ""); err != nil {
		return domain.Task{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.Task{}, err
	}
	return saved, nil
}
func taskWriteRole(role string) bool {
	return role == "owner" || role == "administrator" || role == "marketing_manager" || role == "editor"
}

func lockTask(ctx context.Context, tx pgx.Tx, a domain.Actor, id string) (domain.Task, error) {
	// Rows used for authorization stay locked until the state+history commit.
	var task domain.Task
	var role string
	err := tx.QueryRow(ctx, "SELECT "+taskColumns+",r.role_id,b.status"+taskAccess+` AND t.id=$3 FOR UPDATE OF t FOR SHARE OF b,g,m,r,u,o`, a.OrgID, a.UserID, id).Scan(
		&task.ID, &task.OrgID, &task.BrandID, &task.RequesterID, &task.Title, &task.Objective, &task.Instructions, &task.Priority, &task.Status, &task.AssignedAgentID, &task.Version, &task.DeadlineAt, &task.CreatedAt, &task.UpdatedAt, &role, new(string))
	if err != nil {
		return domain.Task{}, safeErr(err)
	}
	if !taskWriteRole(role) {
		return domain.Task{}, domain.ErrForbidden
	}
	return task, nil
}
func (s *Store) UpdateTask(ctx context.Context, a domain.Actor, in domain.Task) (domain.Task, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return domain.Task{}, err
	}
	defer tx.Rollback(ctx)
	current, err := lockTask(ctx, tx, a, in.ID)
	if err != nil {
		return domain.Task{}, err
	}
	if in.Version < 1 || current.Version != in.Version {
		return domain.Task{}, domain.ErrTaskConflict
	}
	if !current.Editable() || in.BrandID != current.BrandID {
		return domain.Task{}, domain.ErrInvalid
	}
	next := current
	next.Title = strings.TrimSpace(in.Title)
	next.Objective = in.Objective
	next.Instructions = in.Instructions
	next.Priority = in.Priority
	next.DeadlineAt = in.DeadlineAt
	next, err = next.ApplyAssignment(domain.TaskAssignment{AgentID: in.AssignedAgentID, Version: in.Version})
	if err != nil {
		return domain.Task{}, err
	}
	saved, err := saveTask(ctx, tx, a, next)
	if err != nil {
		return domain.Task{}, err
	}
	if err = recordTaskEvent(ctx, tx, a, saved, "updated", current.Status, ""); err != nil {
		return domain.Task{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.Task{}, err
	}
	return saved, nil
}
func (s *Store) TransitionTask(ctx context.Context, a domain.Actor, id string, in domain.TaskTransition) (domain.Task, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return domain.Task{}, err
	}
	defer tx.Rollback(ctx)
	current, err := lockTask(ctx, tx, a, id)
	if err != nil {
		return domain.Task{}, err
	}
	next, err := current.ApplyTransition(in)
	if err != nil {
		return domain.Task{}, err
	}
	saved, err := saveTask(ctx, tx, a, next)
	if err != nil {
		return domain.Task{}, err
	}
	if err = recordTaskEvent(ctx, tx, a, saved, "transitioned", current.Status, in.Reason); err != nil {
		return domain.Task{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.Task{}, err
	}
	return saved, nil
}
func saveTask(ctx context.Context, tx pgx.Tx, a domain.Actor, in domain.Task) (domain.Task, error) {
	return scanTask(tx.QueryRow(ctx, `UPDATE tasks AS t SET title=$4,objective=$5,instructions=$6,priority=$7,status=$8,agent_id=$9,deadline_at=$10,version=version+1,updated_at=now()
 WHERE t.org_id=$1 AND t.id=$2 AND t.version=$3 RETURNING `+taskColumns, a.OrgID, in.ID, in.Version, in.Title, in.Objective, in.Instructions, in.Priority, in.Status, in.AssignedAgentID, in.DeadlineAt))
}
func recordTaskEvent(ctx context.Context, tx pgx.Tx, a domain.Actor, task domain.Task, event string, from domain.TaskStatus, reason string) error {
	_, err := tx.Exec(ctx, `INSERT INTO task_events(org_id,id,task_id,actor_id,event,from_status,to_status,version,reason) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, a.OrgID, uuid.NewString(), task.ID, a.UserID, event, from, task.Status, task.Version, reason)
	if err != nil {
		return err
	}
	return audit(ctx, tx, a, "task."+event, task.ID)
}
func (s *Store) TaskEvents(ctx context.Context, a domain.Actor, id string) ([]domain.TaskEvent, error) {
	// The visibility predicate is repeated at query time, not just in the use case.
	rows, err := s.Pool.Query(ctx, `SELECT e.id,e.task_id,e.actor_id,e.event,e.from_status,e.to_status,e.version,e.reason,e.created_at FROM task_events e
 WHERE e.org_id=$1 AND e.task_id=$3 AND EXISTS(SELECT 1`+taskAccess+` AND t.id=$3) ORDER BY e.version LIMIT 1000`, a.OrgID, a.UserID, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.TaskEvent, 0)
	for rows.Next() {
		var event domain.TaskEvent
		if err = rows.Scan(&event.ID, &event.TaskID, &event.ActorID, &event.Event, &event.FromStatus, &event.ToStatus, &event.Version, &event.Reason, &event.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, event)
	}
	if err = rows.Err(); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	return out, nil
}
