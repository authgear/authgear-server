ALTER TABLE _auth_principal RENAME COLUMN provider TO type;
ALTER TABLE _auth_principal RENAME TO _auth_identity;

ALTER TABLE _auth_provider_oauth RENAME COLUMN principal_id TO identity_id;
ALTER TABLE _auth_provider_oauth RENAME TO _auth_identity_oauth;

ALTER TABLE _auth_provider_password RENAME COLUMN principal_id TO identity_id;
ALTER TABLE _auth_provider_password RENAME TO _auth_identity_login_id;

DROP TABLE _auth_provider_custom_token;
