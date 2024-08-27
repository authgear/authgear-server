-- +migrate Up
ALTER TABLE _auth_identity_ldap ADD COLUMN last_login_username TEXT;

-- +migrate Down
ALTER TABLE _auth_identity_ldap DROP COLUMN last_login_username;
