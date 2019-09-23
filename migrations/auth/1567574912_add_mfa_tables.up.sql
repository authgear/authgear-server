CREATE TABLE _auth_authenticator (
  id TEXT PRIMARY KEY,
  type TEXT NOT NULL,
  user_id TEXT NOT NULL REFERENCES _core_user(id),
  app_id TEXT NOT NULL
);

CREATE TABLE _auth_authenticator_totp (
  id TEXT PRIMARY KEY REFERENCES _auth_authenticator(id),
  activated BOOLEAN NOT NULL,
  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  activated_at TIMESTAMP WITHOUT TIME ZONE,

  secret TEXT NOT NULL,
  display_name TEXT NOT NULL,

  app_id TEXT NOT NULL
);

CREATE TABLE _auth_authenticator_oob (
  id TEXT PRIMARY KEY REFERENCES _auth_authenticator(id),
  activated BOOLEAN NOT NULL,
  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  activated_at TIMESTAMP WITHOUT TIME ZONE,

  channel TEXT NOT NULL,
  phone TEXT NOT NULL,
  email TEXT NOT NULL,

  app_id TEXT NOT NULL
);

CREATE TABLE _auth_authenticator_oob_code (
  id TEXT PRIMARY KEY,
  authenticator_id TEXT NOT NULL REFERENCES _auth_authenticator(id),
  code TEXT NOT NULL,
  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  expire_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,

  app_id TEXT NOT NULL
);

CREATE TABLE _auth_authenticator_recovery_code (
  id TEXT PRIMARY KEY,
  code TEXT NOT NULL,
  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  consumed BOOLEAN NOT NULL,

  app_id TEXT NOT NULL
);

CREATE TABLE _auth_authenticator_bearer_token (
  id TEXT PRIMARY KEY,
  parent_id TEXT NOT NULL REFERENCES _auth_authenticator(id),
  token TEXT NOT NULL,
  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  expire_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,

  app_id TEXT NOT NULL
);
