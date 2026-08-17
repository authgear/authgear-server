-- +migrate Up

CREATE TABLE _auth_oauth_client (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  client_id text NOT NULL,
  -- 'DCR' | 'CIMD'. Matches model.OAuthClientSource verbatim.
  --
  -- NOTE: this table does NOT contain every OAuth client of a project.
  -- Statically configured clients (model.OAuthClientSource 'STATIC') live in
  -- authgear.yaml under oauth.clients and never appear here, so 'STATIC' is
  -- currently not a value this column takes. Any query intended to cover all
  -- clients must also read the project's authgear.yaml.
  source text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  updated_at timestamp without time zone NOT NULL,
  -- CIMD only: timestamp of the most recent successful metadata fetch.
  -- Always NULL for source = 'DCR'.
  last_fetched_at timestamp without time zone,
  kind text NOT NULL,
  application_type text NOT NULL,
  client_name text,
  client_uri text,
  logo_uri text,
  tos_uri text,
  policy_uri text,
  redirect_uris jsonb NOT NULL,
  grant_types jsonb NOT NULL,
  response_types jsonb NOT NULL
);
CREATE UNIQUE INDEX _auth_oauth_client_client_id_unique ON _auth_oauth_client USING btree (app_id, client_id);
CREATE INDEX _auth_oauth_client_app_id_created_at ON _auth_oauth_client USING btree (app_id, created_at);
CREATE INDEX _auth_oauth_client_app_id_source ON _auth_oauth_client USING btree (app_id, source);

-- +migrate Down

DROP INDEX IF EXISTS _auth_oauth_client_app_id_source;
DROP INDEX IF EXISTS _auth_oauth_client_app_id_created_at;
DROP INDEX IF EXISTS _auth_oauth_client_client_id_unique;
DROP TABLE IF EXISTS _auth_oauth_client;
