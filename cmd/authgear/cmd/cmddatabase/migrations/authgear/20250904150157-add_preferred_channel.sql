-- +migrate Up
ALTER TABLE _auth_authenticator_oob ADD COLUMN preferred_channel text NULL;

-- +migrate Down
ALTER TABLE _auth_authenticator_oob DROP COLUMN preferred_channel;
