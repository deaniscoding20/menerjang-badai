-- Append-only human feedback on one exact output version. No dispatch, approval,
-- retry, workflow revision or lifecycle transition is implied.
ALTER TABLE task_deliverable_versions ADD CONSTRAINT task_deliverable_versions_target_key
 UNIQUE(org_id,task_id,id,revision);
CREATE TABLE task_revision_requests (
 org_id uuid NOT NULL,
 id uuid NOT NULL,
 task_id uuid NOT NULL,
 deliverable_id uuid NOT NULL,
 deliverable_revision bigint NOT NULL CHECK(deliverable_revision>0),
 author_id uuid NOT NULL,
 body text NOT NULL CHECK(length(body) BETWEEN 1 AND 4000 AND length(btrim(body))>0),
 created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(org_id,id),
 FOREIGN KEY(org_id,task_id) REFERENCES tasks(org_id,id),
 FOREIGN KEY(org_id,task_id,deliverable_id,deliverable_revision)
  REFERENCES task_deliverable_versions(org_id,task_id,id,revision),
 FOREIGN KEY(org_id,author_id) REFERENCES organization_members(org_id,user_id)
);
CREATE INDEX task_revision_requests_task_time ON task_revision_requests(org_id,task_id,created_at DESC,id DESC);
CREATE FUNCTION reject_task_revision_request_mutation() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 RAISE EXCEPTION 'Task revision requests are immutable' USING ERRCODE='23514';
END;
$$;
CREATE TRIGGER task_revision_requests_immutable BEFORE UPDATE OR DELETE ON task_revision_requests
 FOR EACH ROW EXECUTE FUNCTION reject_task_revision_request_mutation();
ALTER TABLE task_events DROP CONSTRAINT task_events_event_check;
ALTER TABLE task_events ADD CONSTRAINT task_events_event_check CHECK(event IN (
 'created','updated','transitioned','dependency_added','dependency_removed','deliverable_submitted','revision_requested'
));
ALTER TABLE audit_logs DROP CONSTRAINT audit_logs_event_check;
ALTER TABLE audit_logs ADD CONSTRAINT audit_logs_event_check CHECK(event IN (
 'login.success','login.failure','logout','session.revoked','brand.created','brand.updated','brand.archived','identity.provisioned',
 'task.created','task.updated','task.transitioned','task.comment_added','task.dependency_added','task.dependency_removed','task.deliverable_submitted','task.revision_requested'
));
