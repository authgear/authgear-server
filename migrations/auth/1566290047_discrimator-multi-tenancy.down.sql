ALTER TABLE _auth_password_history DROP COLUMN app_id;

ALTER TABLE _auth_principal DROP COLUMN app_id;

ALTER TABLE _auth_provider_custom_token DROP CONSTRAINT _auth_provider_custom_token_token_principal_id_key;
ALTER TABLE _auth_provider_custom_token ADD CONSTRAINT _auth_provider_custom_token_token_principal_id_key UNIQUE(token_principal_id);
ALTER TABLE _auth_provider_custom_token DROP COLUMN app_id;

ALTER TABLE _auth_provider_oauth DROP CONSTRAINT _auth_provider_oauth_provider_user_id;
ALTER TABLE _auth_provider_oauth ADD CONSTRAINT _auth_provider_oauth_provider_user_id UNIQUE(provider_type, provider_keys, provider_user_id);
ALTER TABLE _auth_provider_oauth DROP COLUMN app_id;

ALTER TABLE _auth_provider_password DROP CONSTRAINT _auth_provider_password_login_id_realm;
ALTER TABLE _auth_provider_password ADD CONSTRAINT _auth_provider_password_login_id_realm UNIQUE(login_id, realm);
ALTER TABLE _auth_provider_password DROP COLUMN app_id;

ALTER TABLE _auth_user_profile DROP COLUMN app_id;

ALTER TABLE _auth_verify_code DROP COLUMN app_id;
