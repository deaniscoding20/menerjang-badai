-- Phase 3 development subset: durable preparation, assignment and history only.
-- No worker/dispatch state is synthesized by this migration.
CREATE TABLE tasks (
 org_id uuid NOT NULL,
 id uuid NOT NULL,
 brand_id uuid NOT NULL,
 requester_id uuid NOT NULL,
 title text NOT NULL CHECK(length(btrim(title)) BETWEEN 1 AND 200),
 objective text NOT NULL DEFAULT '' CHECK(length(objective)<=4000),
 instructions text NOT NULL DEFAULT '' CHECK(length(instructions)<=12000),
 priority text NOT NULL DEFAULT 'normal' CHECK(priority IN ('low','normal','high','urgent')),
 status text NOT NULL DEFAULT 'draft' CHECK(status IN ('draft','ready','assigned','queued','running','waiting_input','waiting_approval','completed','failed','cancelled')),
 agent_id text NOT NULL DEFAULT '' CHECK(agent_id IN ('','ec0b8d12-e31e-43a6-9dc3-82a2ee8a1d01','7e66ec1f-afec-4bad-b0eb-b24e05e5de02','07a4226e-5eba-4d2a-95a5-78389ed40603','9b58c0d9-8300-4a10-883d-04bdba2e1b04','cc9f1488-e14c-4203-8f01-9fc9b20ae505','0b1e02ba-94ae-443d-a29c-c84b95019d06')),
 version bigint NOT NULL DEFAULT 1 CHECK(version>0),
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(org_id,id),
 FOREIGN KEY(org_id,brand_id) REFERENCES brands(org_id,id),
 FOREIGN KEY(org_id,requester_id) REFERENCES organization_members(org_id,user_id),
 CHECK(status<>'assigned' OR agent_id<>'')
);
CREATE INDEX tasks_brand_updated ON tasks(org_id,brand_id,updated_at DESC,id);
CREATE TABLE task_events (
 org_id uuid NOT NULL,
 id uuid NOT NULL,
 task_id uuid NOT NULL,
 actor_id uuid NOT NULL,
 event text NOT NULL CHECK(event IN ('created','updated','transitioned')),
 from_status text NOT NULL DEFAULT '',
 to_status text NOT NULL,
 version bigint NOT NULL CHECK(version>0),
 reason text NOT NULL DEFAULT '' CHECK(length(reason)<=2000),
 created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(org_id,id),
 UNIQUE(org_id,task_id,version),
 FOREIGN KEY(org_id,task_id) REFERENCES tasks(org_id,id),
 FOREIGN KEY(org_id,actor_id) REFERENCES organization_members(org_id,user_id)
);
CREATE INDEX task_events_task_time ON task_events(org_id,task_id,version);
ALTER TABLE audit_logs DROP CONSTRAINT audit_logs_event_check;
ALTER TABLE audit_logs ADD CONSTRAINT audit_logs_event_check CHECK(event IN (
 'login.success','login.failure','logout','session.revoked','brand.created','brand.updated','brand.archived','identity.provisioned',
 'task.created','task.updated','task.transitioned'
));
INSERT INTO permissions VALUES ('tasks.read'),('tasks.write');
INSERT INTO role_permissions SELECT id,'tasks.read' FROM roles;
INSERT INTO role_permissions SELECT id,'tasks.write' FROM roles WHERE id IN ('owner','administrator','marketing_manager','editor');
