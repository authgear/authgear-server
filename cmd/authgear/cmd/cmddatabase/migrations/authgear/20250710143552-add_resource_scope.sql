-- +migrate Up

CREATE TABLE _auth_resource (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  updated_at timestamp without time zone NOT NULL,
  uri text NOT NULL,
  name text,
  metadata jsonb
);
CREATE UNIQUE INDEX _auth_resource_uri_unique ON _auth_resource USING btree (app_id, uri);
CREATE INDEX _auth_resource_uri_typeahead ON _auth_resource USING btree (app_id, uri text_pattern_ops);
CREATE INDEX _auth_resource_name_typeahead ON _auth_resource USING btree (app_id, name text_pattern_ops);
CREATE INDEX _auth_resource_app_id_created_at ON _auth_resource USING btree (app_id, created_at);

CREATE TABLE _auth_resource_scope (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  updated_at timestamp without time zone NOT NULL,
  resource_id text NOT NULL REFERENCES _auth_resource(id),
  scope text NOT NULL,
  description text,
  metadata jsonb
);
CREATE UNIQUE INDEX _auth_resource_scope_unique ON _auth_resource_scope USING btree (app_id, resource_id, scope);
CREATE INDEX _auth_resource_scope_scope_typeahead ON _auth_resource_scope USING btree (app_id, resource_id, scope text_pattern_ops);
CREATE INDEX _auth_resource_scope_app_id_resource_id_created_at ON _auth_resource_scope USING btree (app_id, resource_id, created_at);

CREATE TABLE _auth_client_resource (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  updated_at timestamp without time zone NOT NULL,
  client_id text NOT NULL,
  resource_id text NOT NULL REFERENCES _auth_resource(id)
);
CREATE UNIQUE INDEX _auth_client_resource_unique ON _auth_client_resource USING btree (app_id, client_id, resource_id);

CREATE TABLE _auth_client_resource_scope (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  updated_at timestamp without time zone NOT NULL,
  client_id text NOT NULL,
  resource_id text NOT NULL REFERENCES _auth_resource(id),
  scope_id text NOT NULL REFERENCES _auth_resource_scope(id)
);
CREATE UNIQUE INDEX _auth_client_resource_scope_unique ON _auth_client_resource_scope USING btree (app_id, client_id, resource_id, scope_id);

-- +migrate Down

DROP INDEX IF EXISTS _auth_client_resource_scope_unique;
DROP TABLE IF EXISTS _auth_client_resource_scope;
DROP INDEX IF EXISTS _auth_client_resource_unique;
DROP TABLE IF EXISTS _auth_client_resource;
DROP INDEX IF EXISTS _auth_resource_scope_app_id_resource_id_created_at;
DROP INDEX IF EXISTS _auth_resource_scope_scope_typeahead;
DROP INDEX IF EXISTS _auth_resource_scope_unique;
DROP TABLE IF EXISTS _auth_resource_scope;
DROP INDEX IF EXISTS _auth_resource_app_id_created_at;
DROP INDEX IF EXISTS _auth_resource_name_typeahead;
DROP INDEX IF EXISTS _auth_resource_uri_typeahead;
DROP INDEX IF EXISTS _auth_resource_uri_unique;
DROP TABLE IF EXISTS _auth_resource;
