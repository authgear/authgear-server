ALTER TABLE _auth_password_history ADD COLUMN app_id TEXT;
UPDATE _auth_password_history SET app_id = '';
ALTER TABLE _auth_password_history ALTER COLUMN app_id SET NOT NULL;
CREATE INDEX _auth_password_history_app_id_idx ON _auth_password_history(app_id);

ALTER TABLE _auth_principal ADD COLUMN app_id TEXT;
UPDATE _auth_principal SET app_id = '';
ALTER TABLE _auth_principal ALTER COLUMN app_id SET NOT NULL;
CREATE INDEX _auth_principal_app_id_idx ON _auth_principal(app_id);

ALTER TABLE _auth_provider_custom_token ADD COLUMN app_id TEXT;
UPDATE _auth_provider_custom_token SET app_id = '';
ALTER TABLE _auth_provider_custom_token ALTER COLUMN app_id SET NOT NULL;
CREATE INDEX _auth_provider_custom_token_app_id_idx ON _auth_provider_custom_token(app_id);
ALTER TABLE _auth_provider_custom_token DROP CONSTRAINT _auth_provider_custom_token_token_principal_id_key;
ALTER TABLE _auth_provider_custom_token ADD CONSTRAINT _auth_provider_custom_token_token_principal_id_key UNIQUE(app_id, token_principal_id);

ALTER TABLE _auth_provider_oauth ADD COLUMN app_id TEXT;
UPDATE _auth_provider_oauth SET app_id = '';
ALTER TABLE _auth_provider_oauth ALTER COLUMN app_id SET NOT NULL;
CREATE INDEX _auth_provider_oauth_app_id_idx ON _auth_provider_oauth(app_id);
ALTER TABLE _auth_provider_oauth DROP CONSTRAINT _auth_provider_oauth_provider_user_id;
ALTER TABLE _auth_provider_oauth ADD CONSTRAINT _auth_provider_oauth_provider_user_id UNIQUE(app_id, provider_type, provider_keys, provider_user_id);

ALTER TABLE _auth_provider_password ADD COLUMN app_id TEXT;
UPDATE _auth_provider_password SET app_id = '';
ALTER TABLE _auth_provider_password ALTER COLUMN app_id SET NOT NULL;
CREATE INDEX _auth_provider_password_app_id_idx ON _auth_provider_password(app_id);
ALTER TABLE _auth_provider_password DROP CONSTRAINT _auth_provider_password_login_id_realm;
ALTER TABLE _auth_provider_password ADD CONSTRAINT _auth_provider_password_login_id_realm UNIQUE(app_id, login_id, realm);

ALTER TABLE _auth_user_profile ADD COLUMN app_id TEXT;
UPDATE _auth_user_profile SET app_id = '';
ALTER TABLE _auth_user_profile ALTER COLUMN app_id SET NOT NULL;
CREATE INDEX _auth_user_profile_app_id_idx ON _auth_user_profile(app_id);

ALTER TABLE _auth_verify_code ADD COLUMN app_id TEXT;
UPDATE _auth_verify_code SET app_id = '';
ALTER TABLE _auth_verify_code ALTER COLUMN app_id SET NOT NULL;
CREATE INDEX _auth_verify_code_app_id_idx ON _auth_verify_code(app_id);
