-- +migrate Up
ALTER TABLE _auth_user ADD COLUMN is_indefinitely_disabled boolean;
ALTER TABLE _auth_user ADD COLUMN account_valid_from timestamp without time zone;
ALTER TABLE _auth_user ADD COLUMN account_valid_until timestamp without time zone;
ALTER TABLE _auth_user ADD COLUMN temporarily_disabled_from timestamp without time zone;
ALTER TABLE _auth_user ADD COLUMN temporarily_disabled_until timestamp without time zone;
ALTER TABLE _auth_user ADD COLUMN account_status_stale_from timestamp without time zone;
CREATE INDEX _auth_user_account_status_stale_from ON _auth_user (account_status_stale_from);

-- +migrate Down
ALTER TABLE _auth_user DROP COLUMN is_indefinitely_disabled;
ALTER TABLE _auth_user DROP COLUMN account_valid_from;
ALTER TABLE _auth_user DROP COLUMN account_valid_until;
ALTER TABLE _auth_user DROP COLUMN temporarily_disabled_from;
ALTER TABLE _auth_user DROP COLUMN temporarily_disabled_until;
ALTER TABLE _auth_user DROP COLUMN account_status_stale_from;
