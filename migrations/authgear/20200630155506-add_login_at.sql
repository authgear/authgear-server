-- +migrate Up
ALTER TABLE _auth_user ADD COLUMN login_at timestamp without time zone;
UPDATE _auth_user SET login_at = last_login_at;

-- +migrate Down
ALTER TABLE _auth_user DROP COLUMN login_at;
