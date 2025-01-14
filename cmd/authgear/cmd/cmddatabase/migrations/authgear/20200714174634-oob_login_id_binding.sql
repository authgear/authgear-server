-- +migrate Up
ALTER TABLE _auth_authenticator_oob ADD COLUMN identity_id text references _auth_identity;

-- +migrate Down
ALTER TABLE _auth_authenticator_oob DROP COLUMN identity_id;
