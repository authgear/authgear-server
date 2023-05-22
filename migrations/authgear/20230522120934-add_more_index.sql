-- +migrate Up
CREATE INDEX _auth_authenticator_app_id_user_id ON _auth_authenticator (app_id, user_id);
CREATE INDEX _auth_authenticator_app_id ON _auth_authenticator (app_id);
CREATE INDEX _auth_identity_app_id_user_id ON _auth_identity (app_id, user_id);
CREATE INDEX _auth_identity_app_id ON _auth_identity (app_id);
CREATE INDEX _auth_oauth_authorization_app_id_user_id ON _auth_oauth_authorization (app_id, user_id);
CREATE INDEX _auth_password_history_app_id_user_id ON _auth_password_history (app_id, user_id);
CREATE INDEX _auth_recovery_code_app_id_user_id ON _auth_recovery_code (app_id, user_id);
CREATE INDEX _auth_verified_claim_app_id_user_id ON _auth_verified_claim (app_id, user_id);
CREATE INDEX _auth_identity_login_id_login_id ON _auth_identity_login_id (app_id, login_id_key, login_id);
CREATE INDEX _auth_identity_login_id_claim_email ON _auth_identity_login_id (app_id, (claims ->> 'email'));
CREATE INDEX _auth_identity_login_id_claim_phone_number ON _auth_identity_login_id (app_id, (claims ->> 'phone_number'));
CREATE INDEX _auth_identity_oauth_claim_email ON _auth_identity_oauth (app_id, (claims ->> 'email'));
CREATE INDEX _auth_identity_oauth_claim_phone_number ON _auth_identity_oauth (app_id, (claims ->> 'phone_number'));

-- +migrate Down
DROP INDEX _auth_authenticator_app_id_user_id;
DROP INDEX _auth_authenticator_app_id;
DROP INDEX _auth_identity_app_id_user_id;
DROP INDEX _auth_identity_app_id;
DROP INDEX _auth_oauth_authorization_app_id_user_id;
DROP INDEX _auth_password_history_app_id_user_id;
DROP INDEX _auth_recovery_code_app_id_user_id;
DROP INDEX _auth_verified_claim_app_id_user_id;
DROP INDEX _auth_identity_login_id_login_id;
DROP INDEX _auth_identity_login_id_claim_email;
DROP INDEX _auth_identity_login_id_claim_phone_number;
DROP INDEX _auth_identity_oauth_claim_email;
DROP INDEX _auth_identity_oauth_claim_phone_number;
