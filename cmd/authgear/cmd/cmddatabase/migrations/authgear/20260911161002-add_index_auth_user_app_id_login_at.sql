-- +migrate Up
-- The Admin API sorts users by their most recent login, which _auth_user
-- stores in login_at (last_login_at holds the login before that). Index the
-- column actually sorted on, and drop the index on the column no longer used
-- for sorting.
CREATE INDEX _auth_user_app_id_login_at ON _auth_user (app_id, login_at DESC NULLS LAST);
DROP INDEX _auth_user_app_id_last_login_at;

-- +migrate Down
CREATE INDEX _auth_user_app_id_last_login_at ON _auth_user (app_id, last_login_at DESC NULLS LAST);
DROP INDEX _auth_user_app_id_login_at;
