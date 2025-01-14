-- +migrate Up

ALTER TABLE _auth_user ADD COLUMN is_disabled boolean;
UPDATE _auth_user SET is_disabled = FALSE;
ALTER TABLE _auth_user ALTER COLUMN is_disabled SET NOT NULL;

ALTER TABLE _auth_user ADD COLUMN disable_reason text;

-- +migrate Down

ALTER TABLE _auth_user DROP COLUMN is_disabled;
ALTER TABLE _auth_user DROP COLUMN disable_reason;
