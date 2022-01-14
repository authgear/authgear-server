-- +migrate Up
ALTER TABLE _auth_user ADD COLUMN custom_attributes jsonb;

-- +migrate Down
ALTER TABLE _auth_user DROP COLUMN custom_attributes;
