-- +migrate Up
-- The Users list sorts on login_at (most recent login), not last_login_at (the one before it).
CREATE INDEX _auth_user_app_id_login_at ON _auth_user (app_id, login_at DESC NULLS LAST);

-- SortOption.Apply always emits NULLS LAST, which a DESC NULLS LAST index cannot
-- serve ascending, so the ascending sorts need their own index.
CREATE INDEX _auth_user_app_id_login_at_asc ON _auth_user (app_id, login_at ASC NULLS LAST);
CREATE INDEX _auth_user_app_id_created_at_asc ON _auth_user (app_id, created_at ASC NULLS LAST);

-- _auth_user_app_id_last_login_at is kept so a rolled-back binary still sorts on an
-- index. Drop it in a later release.

-- +migrate Down
DROP INDEX _auth_user_app_id_created_at_asc;
DROP INDEX _auth_user_app_id_login_at_asc;
DROP INDEX _auth_user_app_id_login_at;
