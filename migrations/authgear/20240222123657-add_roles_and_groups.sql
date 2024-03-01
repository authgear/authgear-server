-- +migrate Up
CREATE TABLE _auth_role (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  updated_at timestamp without time zone NOT NULL,
  key text NOT NULL,
  name text,
  description text
);
-- Each project has its own set of roles. The role keys are unique within a project.
CREATE UNIQUE INDEX _auth_role_key_unique ON _auth_role USING btree (app_id, key);
-- This index supports typeahead search for roles within a project.
CREATE INDEX _auth_role_key_typeahead ON _auth_role USING btree (app_id, key text_pattern_ops);
CREATE INDEX _auth_role_name_typeahead ON _auth_role USING btree (app_id, name text_pattern_ops);

CREATE TABLE _auth_group (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  updated_at timestamp without time zone NOT NULL,
  key text NOT NULL,
  name text,
  description text
);
-- Each project has its own set of groups. The group keys are unique within a project.
CREATE UNIQUE INDEX _auth_group_key_unique ON _auth_group USING btree (app_id, key);
-- This index supports typeahead search for groups within a project.
CREATE INDEX _auth_group_key_typeahead ON _auth_group USING btree (app_id, key text_pattern_ops);
CREATE INDEX _auth_group_name_typeahead ON _auth_group USING btree (app_id, name text_pattern_ops);

CREATE TABLE _auth_group_role (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  updated_at timestamp without time zone NOT NULL,
  group_id text NOT NULL REFERENCES _auth_group(id),
  role_id text NOT NULL REFERENCES _auth_role(id)
);
-- A role and a group can only be associated at most once.
CREATE UNIQUE INDEX _auth_group_role_unique ON _auth_group_role USING btree (app_id, group_id, role_id);
-- This index supports joining from Group.
CREATE INDEX _auth_group_role_group ON _auth_group_role USING btree (app_id, group_id);
-- This index supports joining from Role.
CREATE INDEX _auth_group_role_role ON _auth_group_role USING btree (app_id, role_id);

CREATE TABLE _auth_user_role (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  updated_at timestamp without time zone NOT NULL,
  user_id text NOT NULL REFERENCES _auth_user(id),
  role_id text NOT NULL REFERENCES _auth_role(id)
);
-- A role and a user can only be associated at most once.
CREATE UNIQUE INDEX _auth_user_role_unique ON _auth_user_role USING btree (app_id, user_id, role_id);
-- This index supports joining from User.
CREATE INDEX _auth_user_role_user ON _auth_user_role USING btree (app_id, user_id);
-- This index supports joining from Role.
CREATE INDEX _auth_user_role_role ON _auth_user_role USING btree (app_id, role_id);

CREATE TABLE _auth_user_group (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  updated_at timestamp without time zone NOT NULL,
  user_id text NOT NULL REFERENCES _auth_user(id),
  group_id text NOT NULL REFERENCES _auth_group(id)
);
-- A group and a user can only be associated at most once.
CREATE UNIQUE INDEX _auth_user_group_unique ON _auth_user_group USING btree(app_id, user_id, group_id);
-- This index supports joining from User.
CREATE INDEX _auth_user_group_user ON _auth_user_group USING btree (app_id, user_id);
-- This index supports joining from Role.
CREATE INDEX _auth_user_group_group ON _auth_user_group USING btree (app_id, group_id);

-- +migrate Down
DROP TABLE _auth_user_group;
DROP TABLE _auth_user_role;
DROP TABLE _auth_group_role;
DROP TABLE _auth_group;
DROP TABLE _auth_role;
