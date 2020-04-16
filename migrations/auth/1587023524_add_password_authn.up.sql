CREATE TABLE _auth_authenticator_password (
  id TEXT PRIMARY KEY REFERENCES _auth_authenticator(id),
  app_id TEXT NOT NULL,
  password_hash TEXT NOT NULL
);

INSERT INTO _auth_authenticator(id, type, user_id, app_id)
    SELECT _auth_identity_login_id.identity_id,
           'password',
           _auth_identity.user_id,
           _auth_identity.app_id
        FROM _auth_identity_login_id
        JOIN _auth_identity ON (_auth_identity_login_id.identity_id = _auth_identity.id);

INSERT INTO _auth_authenticator_password(id, app_id, password_hash)
    SELECT identity_id, app_id, password
        FROM _auth_identity_login_id;
