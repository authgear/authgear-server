ALTER TABLE _auth_provider_password ADD COLUMN realm TEXT;
UPDATE _auth_provider_password SET realm = 'default';
ALTER TABLE _auth_provider_password ALTER COLUMN realm SET NOT NULL;

ALTER TABLE _auth_provider_password DROP CONSTRAINT _auth_provider_password_login_id_key;
ALTER TABLE _auth_provider_password ADD CONSTRAINT  _auth_provider_password_login_id_realm UNIQUE (login_id, realm);
