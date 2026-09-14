-- +migrate Up
-- The Admin API sorts users by their most recent login, which _auth_user
-- stores in login_at (last_login_at holds the login before that). Index the
-- column actually sorted on.
CREATE INDEX _auth_user_app_id_login_at ON _auth_user (app_id, login_at DESC NULLS LAST);

-- SortOption.Apply always emits NULLS LAST, in both directions. A btree can
-- only be scanned forwards or backwards, so (col DESC NULLS LAST) serves
-- "DESC NULLS LAST" and "ASC NULLS FIRST" -- never "ASC NULLS LAST". Sorting
-- the Users list ascending therefore fell back to a sequential scan of the
-- whole table plus a sort, on every page. Index the ascending order too, for
-- login_at and for created_at, which has the same gap.
-- (QueryForExport sidesteps this by ordering ASC NULLS FIRST on purpose; the
-- list cannot, because that would move never-logged-in users to the top.)
CREATE INDEX _auth_user_app_id_login_at_asc ON _auth_user (app_id, login_at ASC NULLS LAST);
CREATE INDEX _auth_user_app_id_created_at_asc ON _auth_user (app_id, created_at ASC NULLS LAST);

-- _auth_user_app_id_last_login_at is left in place on purpose. Nothing sorts
-- by last_login_at any more, but a deployment that applies this migration and
-- then rolls the binary back would have the previous code sort on an
-- unindexed column, turning every Users list page into a sequential scan of
-- _auth_user. Drop it in a later release, once rolling back that far is no
-- longer a concern.

-- +migrate Down
DROP INDEX _auth_user_app_id_created_at_asc;
DROP INDEX _auth_user_app_id_login_at_asc;
DROP INDEX _auth_user_app_id_login_at;
