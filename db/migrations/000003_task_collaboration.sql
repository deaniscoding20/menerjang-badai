-- Append-only human notes and optional deadline metadata; no runtime execution.
ALTER TABLE tasks ADD COLUMN deadline_at timestamptz;
CREATE INDEX tasks_deadline ON tasks(org_id,brand_id,deadline_at) WHERE deadline_at IS NOT NULL;
CREATE TABLE task_comments (
 org_id uuid NOT NULL,
 id uuid NOT NULL,
 task_id uuid NOT NULL,
 author_id uuid NOT NULL,
 kind text NOT NULL DEFAULT 'comment' CHECK(kind IN ('comment','clarification')),
 body text NOT NULL CHECK(length(btrim(body)) BETWEEN 1 AND 4000),
 created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(org_id,id),
 FOREIGN KEY(org_id,task_id) REFERENCES tasks(org_id,id),
 FOREIGN KEY(org_id,author_id) REFERENCES organization_members(org_id,user_id)
);
CREATE INDEX task_comments_task_time ON task_comments(org_id,task_id,created_at,id);
ALTER TABLE audit_logs DROP CONSTRAINT audit_logs_event_check;
ALTER TABLE audit_logs ADD CONSTRAINT audit_logs_event_check CHECK(event IN (
 'login.success','login.failure','logout','session.revoked','brand.created','brand.updated','brand.archived','identity.provisioned',
 'task.created','task.updated','task.transitioned','task.comment_added'
));
