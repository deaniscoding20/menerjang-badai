package postgres

import (
	"context"
	"github.com/google/uuid"
	"github.com/menerjangbadai/dashboard/internal/application"
	"github.com/menerjangbadai/dashboard/internal/domain"
)

var _ application.TaskCommentRepository = (*Store)(nil)

func (s *Store) AddTaskComment(ctx context.Context, a domain.Actor, id string, in domain.TaskComment) (domain.TaskComment, error) {
	in, err := domain.NewTaskComment(in)
	if err != nil {
		return domain.TaskComment{}, err
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return domain.TaskComment{}, err
	}
	defer tx.Rollback(ctx)
	// Recheck and lock authoritative membership and brand grant at mutation time.
	if _, err = lockTask(ctx, tx, a, id); err != nil {
		return domain.TaskComment{}, err
	}
	saved := domain.TaskComment{}
	err = tx.QueryRow(ctx, `INSERT INTO task_comments(org_id,id,task_id,author_id,kind,body) VALUES($1,$2,$3,$4,$5,$6)
 RETURNING id,task_id,author_id,kind,body,created_at`, a.OrgID, uuid.NewString(), id, a.UserID, in.Kind, in.Body).Scan(&saved.ID, &saved.TaskID, &saved.AuthorID, &saved.Kind, &saved.Body, &saved.CreatedAt)
	if err != nil {
		return domain.TaskComment{}, err
	}
	if err = audit(ctx, tx, a, "task.comment_added", id); err != nil {
		return domain.TaskComment{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.TaskComment{}, err
	}
	return saved, nil
}
func (s *Store) ListTaskComments(ctx context.Context, a domain.Actor, id string) ([]domain.TaskComment, error) {
	rows, err := s.Pool.Query(ctx, `SELECT c.id,c.task_id,c.author_id,c.kind,c.body,c.created_at FROM task_comments c
 WHERE c.org_id=$1 AND c.task_id=$3 AND EXISTS(SELECT 1`+taskAccess+` AND t.id=$3) ORDER BY c.created_at,c.id LIMIT 500`, a.OrgID, a.UserID, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.TaskComment, 0)
	for rows.Next() {
		var comment domain.TaskComment
		if err = rows.Scan(&comment.ID, &comment.TaskID, &comment.AuthorID, &comment.Kind, &comment.Body, &comment.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, comment)
	}
	return out, rows.Err()
}
