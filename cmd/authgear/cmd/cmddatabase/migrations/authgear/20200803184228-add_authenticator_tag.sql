-- +migrate Up
ALTER TABLE _auth_authenticator ADD COLUMN tag JSONB NOT NULL;

-- +migrate Down
ALTER TABLE _auth_authenticator DROP COLUMN tag;
