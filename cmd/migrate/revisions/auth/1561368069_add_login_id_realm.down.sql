ALTER TABLE _auth_provider_password DROP COLUMN realm;

ALTER TABLE _auth_provider_password ADD CONSTRAINT _auth_provider_password_login_id_key UNIQUE (login_id);
