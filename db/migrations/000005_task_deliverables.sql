-- Immutable human-imported output revisions. No agent provenance or approval implied.
CREATE TABLE task_deliverable_versions (
 org_id uuid NOT NULL,
 id uuid NOT NULL,
 task_id uuid NOT NULL,
 revision bigint NOT NULL CHECK(revision>0),
 author_id uuid NOT NULL,
 origin text NOT NULL CHECK(origin='human_import'),
 raw_output text NOT NULL CHECK(octet_length(raw_output) BETWEEN 1 AND 131072),
 validation_status text NOT NULL CHECK(validation_status IN ('validated','invalid')),
 structured jsonb,
 created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(org_id,id),
 UNIQUE(org_id,task_id,revision),
 FOREIGN KEY(org_id,task_id) REFERENCES tasks(org_id,id),
 FOREIGN KEY(org_id,author_id) REFERENCES organization_members(org_id,user_id),
 CHECK((validation_status='validated' AND structured IS NOT NULL AND jsonb_typeof(structured)='object') OR (validation_status='invalid' AND structured IS NULL))
);
CREATE INDEX task_deliverables_revision ON task_deliverable_versions(org_id,task_id,revision DESC);
CREATE FUNCTION reject_deliverable_version_mutation() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 RAISE EXCEPTION 'Deliverable versions are immutable' USING ERRCODE='23514';
END;
$$;
CREATE TRIGGER task_deliverable_versions_immutable BEFORE UPDATE OR DELETE ON task_deliverable_versions
 FOR EACH ROW EXECUTE FUNCTION reject_deliverable_version_mutation();
ALTER TABLE task_events DROP CONSTRAINT task_events_event_check;
ALTER TABLE task_events ADD CONSTRAINT task_events_event_check CHECK(event IN ('created','updated','transitioned','dependency_added','dependency_removed','deliverable_submitted'));
ALTER TABLE audit_logs DROP CONSTRAINT audit_logs_event_check;
ALTER TABLE audit_logs ADD CONSTRAINT audit_logs_event_check CHECK(event IN (
 'login.success','login.failure','logout','session.revoked','brand.created','brand.updated','brand.archived','identity.provisioned',
 'task.created','task.updated','task.transitioned','task.comment_added','task.dependency_added','task.dependency_removed','task.deliverable_submitted'
));
