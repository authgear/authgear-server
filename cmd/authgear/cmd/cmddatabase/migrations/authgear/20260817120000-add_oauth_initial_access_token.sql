-- +migrate Up

CREATE TABLE _auth_oauth_initial_access_token (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  expires_at timestamp without time zone NOT NULL,
  token_type text NOT NULL,
  token_hash text NOT NULL
);
CREATE UNIQUE INDEX _auth_oauth_initial_access_token_hash_unique ON _auth_oauth_initial_access_token USING btree (app_id, token_hash);
CREATE INDEX _auth_oauth_initial_access_token_app_id_created_at ON _auth_oauth_initial_access_token USING btree (app_id, created_at);

-- +migrate Down

DROP INDEX IF EXISTS _auth_oauth_initial_access_token_app_id_created_at;
DROP INDEX IF EXISTS _auth_oauth_initial_access_token_hash_unique;
DROP TABLE IF EXISTS _auth_oauth_initial_access_token;
