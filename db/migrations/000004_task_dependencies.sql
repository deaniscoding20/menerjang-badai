-- Preparation graph only. Execution dispatch remains disabled.
ALTER TABLE tasks ADD CONSTRAINT tasks_org_brand_id_key UNIQUE(org_id,brand_id,id);
CREATE TABLE task_dependencies (
 org_id uuid NOT NULL,
 brand_id uuid NOT NULL,
 task_id uuid NOT NULL,
 dependency_id uuid NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(org_id,task_id,dependency_id),
 CHECK(task_id<>dependency_id),
 FOREIGN KEY(org_id,brand_id,task_id) REFERENCES tasks(org_id,brand_id,id),
 FOREIGN KEY(org_id,brand_id,dependency_id) REFERENCES tasks(org_id,brand_id,id)
);
CREATE INDEX task_dependencies_reverse ON task_dependencies(org_id,brand_id,dependency_id,task_id);
ALTER TABLE task_events DROP CONSTRAINT task_events_event_check;
ALTER TABLE task_events ADD CONSTRAINT task_events_event_check CHECK(event IN ('created','updated','transitioned','dependency_added','dependency_removed'));
ALTER TABLE audit_logs DROP CONSTRAINT audit_logs_event_check;
ALTER TABLE audit_logs ADD CONSTRAINT audit_logs_event_check CHECK(event IN (
 'login.success','login.failure','logout','session.revoked','brand.created','brand.updated','brand.archived','identity.provisioned',
 'task.created','task.updated','task.transitioned','task.comment_added','task.dependency_added','task.dependency_removed'
));
