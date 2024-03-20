-- +migrate Up
CREATE INDEX _auth_authenticator_user_id ON _auth_authenticator (user_id);
CREATE INDEX _auth_group_role_group_id ON _auth_group_role (group_id);
CREATE INDEX _auth_group_role_role_id ON _auth_group_role (role_id);
CREATE INDEX _auth_identity_user_id ON _auth_identity (user_id);
CREATE INDEX _auth_oauth_authorization_user_id ON _auth_oauth_authorization (user_id);
CREATE INDEX _auth_password_history_user_id ON _auth_password_history (user_id);
CREATE INDEX _auth_recovery_code_user_id ON _auth_recovery_code (user_id);
CREATE INDEX _auth_user_group_user_id ON _auth_user_group (user_id);
CREATE INDEX _auth_user_group_group_id ON _auth_user_group (group_id);
CREATE INDEX _auth_user_role_user_id ON _auth_user_role (user_id);
CREATE INDEX _auth_user_role_role_id ON _auth_user_role (role_id);
CREATE INDEX _auth_verified_claim_user_id ON _auth_verified_claim (user_id);

CREATE INDEX _auth_user_app_id ON _auth_user (app_id);
CREATE INDEX _auth_user_app_id_created_at ON _auth_user (app_id, created_at DESC NULLS LAST);
CREATE INDEX _auth_user_app_id_last_login_at ON _auth_user (app_id, last_login_at DESC NULLS LAST);

-- +migrate Down
DROP INDEX _auth_authenticator_user_id;
DROP INDEX _auth_group_role_group_id;
DROP INDEX _auth_group_role_role_id;
DROP INDEX _auth_identity_user_id;
DROP INDEX _auth_oauth_authorization_user_id;
DROP INDEX _auth_password_history_user_id;
DROP INDEX _auth_recovery_code_user_id;
DROP INDEX _auth_user_group_user_id;
DROP INDEX _auth_user_group_group_id;
DROP INDEX _auth_user_role_user_id;
DROP INDEX _auth_user_role_role_id;
DROP INDEX _auth_verified_claim_user_id;

DROP INDEX _auth_user_app_id;
DROP INDEX _auth_user_app_id_created_at;
DROP INDEX _auth_user_app_id_last_login_at;
