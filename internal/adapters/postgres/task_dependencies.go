package postgres

import (
	"context"
	"encoding/json"
	"github.com/menerjangbadai/dashboard/internal/application"
	"github.com/menerjangbadai/dashboard/internal/domain"
)

var _ application.TaskDependencyRepository = (*Store)(nil)

func (s *Store) ListTaskDependencies(ctx context.Context, a domain.Actor, id string) (domain.TaskDependencies, error) {
	// One SQL snapshot binds this dependency list to its source optimistic version.
	var out domain.TaskDependencies
	var raw []byte
	err := s.Pool.QueryRow(ctx, `SELECT t.version,COALESCE((
 SELECT jsonb_agg(jsonb_build_object('task_id',p.id,'title',p.title,'status',p.status) ORDER BY p.title,p.id)
 FROM task_dependencies d JOIN tasks p ON p.org_id=d.org_id AND p.brand_id=d.brand_id AND p.id=d.dependency_id
 WHERE d.org_id=t.org_id AND d.brand_id=t.brand_id AND d.task_id=t.id
 ),'[]'::jsonb)`+taskAccess+` AND t.id=$3`, a.OrgID, a.UserID, id).Scan(&out.Version, &raw)
	if err != nil {
		return domain.TaskDependencies{}, safeErr(err)
	}
	if err = json.Unmarshal(raw, &out.Items); err != nil {
		return domain.TaskDependencies{}, err
	}
	return out, nil
}

func (s *Store) ChangeTaskDependency(ctx context.Context, a domain.Actor, id string, in domain.TaskDependencyChange, add bool) (domain.Task, error) {
	// Only visible tasks may choose a graph lock scope. Never accept client org/brand.
	visible, err := s.GetTask(ctx, a, id)
	if err != nil {
		return domain.Task{}, err
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return domain.Task{}, err
	}
	defer tx.Rollback(ctx)
	// Serialize all graph mutations within this org+brand BEFORE taking task locks.
	// Source-row locks alone let concurrent A->B and B->A both pass cycle checks.
	// Advisory locking avoids reversing lockTask's existing ACL/row lock order.
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,819202304))`, visible.OrgID+":"+visible.BrandID); err != nil {
		return domain.Task{}, err
	}
	current, err := lockTask(ctx, tx, a, id)
	if err != nil {
		return domain.Task{}, err
	}
	if err = current.ValidateDependencyChange(in); err != nil {
		return domain.Task{}, err
	}
	target, err := scanTask(tx.QueryRow(ctx, "SELECT "+taskColumns+taskAccess+` AND t.id=$3 AND t.brand_id=$4 FOR SHARE OF t,b,g,m,r,u,o`, a.OrgID, a.UserID, in.DependencyID, current.BrandID))
	if err != nil {
		return domain.Task{}, err
	}
	var exists bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM task_dependencies WHERE org_id=$1 AND task_id=$2 AND dependency_id=$3)`, a.OrgID, current.ID, target.ID).Scan(&exists); err != nil {
		return domain.Task{}, err
	}
	if exists == add {
		// No-op writes still enforce session role, scope, editability and current version.
		if err = tx.Commit(ctx); err != nil {
			return domain.Task{}, err
		}
		return current, nil
	}
	event := "dependency_removed"
	if add {
		var cycle bool
		err = tx.QueryRow(ctx, `WITH RECURSIVE reachable(id) AS (
   SELECT $3::uuid UNION SELECT d.dependency_id FROM task_dependencies d JOIN reachable r ON d.task_id=r.id
   WHERE d.org_id=$1 AND d.brand_id=$2
  ) SELECT EXISTS(SELECT 1 FROM reachable WHERE id=$4::uuid)`, a.OrgID, current.BrandID, target.ID, current.ID).Scan(&cycle)
		if err != nil {
			return domain.Task{}, err
		}
		if cycle {
			return domain.Task{}, domain.ErrDependencyCycle
		}
		_, err = tx.Exec(ctx, `INSERT INTO task_dependencies(org_id,brand_id,task_id,dependency_id) VALUES($1,$2,$3,$4)`, a.OrgID, current.BrandID, current.ID, target.ID)
		event = "dependency_added"
	} else {
		_, err = tx.Exec(ctx, `DELETE FROM task_dependencies WHERE org_id=$1 AND brand_id=$2 AND task_id=$3 AND dependency_id=$4`, a.OrgID, current.BrandID, current.ID, target.ID)
	}
	if err != nil {
		return domain.Task{}, err
	}
	saved, err := saveTask(ctx, tx, a, current)
	if err != nil {
		return domain.Task{}, err
	}
	if err = recordTaskEvent(ctx, tx, a, saved, event, current.Status, target.ID); err != nil {
		return domain.Task{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.Task{}, err
	}
	return saved, nil
}
