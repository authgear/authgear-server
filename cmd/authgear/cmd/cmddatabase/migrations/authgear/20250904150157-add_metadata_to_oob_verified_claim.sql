-- +migrate Up
ALTER TABLE _auth_authenticator_oob ADD COLUMN metadata jsonb NULL;
ALTER TABLE _auth_verified_claim ADD COLUMN metadata jsonb NULL;

-- +migrate Down
ALTER TABLE _auth_authenticator_oob DROP COLUMN metadata;
ALTER TABLE _auth_verified_claim DROP COLUMN metadata;
