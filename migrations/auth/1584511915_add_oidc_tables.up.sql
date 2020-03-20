CREATE TABLE _auth_oauth_authorization (
  id TEXT PRIMARY KEY,
  app_id TEXT NOT NULL,
  client_id TEXT NOT NULL,
  user_id TEXT NOT NULL REFERENCES _core_user(id),
  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  scopes JSONB NOT NULL,
  UNIQUE (app_id, client_id, user_id)
);
