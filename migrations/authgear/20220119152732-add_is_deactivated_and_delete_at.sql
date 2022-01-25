-- +migrate Up
ALTER TABLE _auth_user ADD COLUMN is_deactivated boolean;
ALTER TABLE _auth_user ADD COLUMN delete_at timestamp without time zone;
CREATE INDEX _auth_user_delete_at ON _auth_user USING BRIN (delete_at);

-- +migrate Down
ALTER TABLE _auth_user DROP COLUMN is_deactivated;
ALTER TABLE _auth_user DROP COLUMN delete_at;
