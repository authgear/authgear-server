-- +migrate Up
ALTER TABLE _auth_user ADD COLUMN standard_attributes jsonb;

-- +migrate Down
ALTER TABLE _auth_user DROP COLUMN standard_attributes;
