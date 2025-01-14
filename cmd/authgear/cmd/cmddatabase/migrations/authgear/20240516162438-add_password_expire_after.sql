-- +migrate Up
ALTER TABLE _auth_authenticator_password ADD COLUMN expire_after timestamp without time zone;

-- +migrate Down
ALTER TABLE _auth_authenticator_password DROP COLUMN expire_after;
