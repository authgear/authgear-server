-- +migrate Up
ALTER TABLE _auth_identity_ldap ADD COLUMN last_login_username TEXT;
UPDATE _auth_identity_ldap
SET last_login_username = '';
ALTER TABLE _auth_identity_ldap ALTER COLUMN last_login_username SET NOT NULL;

-- +migrate Down
ALTER TABLE _auth_identity_ldap DROP COLUMN last_login_username;
