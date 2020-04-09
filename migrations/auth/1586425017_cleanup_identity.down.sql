ALTER TABLE _auth_identity RENAME COLUMN provider TO type;
ALTER TABLE _auth_identity RENAME TO _auth_principal;

ALTER TABLE _auth_identity_oauth RENAME COLUMN identity_id TO principal_id;
ALTER TABLE _auth_identity_oauth RENAME TO _auth_provider_oauth;

ALTER TABLE _auth_identity_login_id RENAME COLUMN identity_id TO principal_id;
ALTER TABLE _auth_identity_login_id RENAME TO _auth_provider_password;

CREATE TABLE _auth_provider_custom_token (
    principal_id text REFERENCES _auth_principal(id) PRIMARY KEY,
    token_principal_id text NOT NULL,
    raw_profile jsonb NOT NULL,
    claims jsonb NOT NULL,
    app_id text NOT NULL,
    UNIQUE (app_id, token_principal_id)
);
CREATE INDEX _auth_provider_custom_token_app_id_idx ON _auth_provider_custom_token(app_id);
