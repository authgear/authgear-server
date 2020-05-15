ALTER TABLE _auth_identity_login_id DROP COLUMN password;
ALTER TABLE _auth_identity_login_id DROP COLUMN realm;
UPDATE _auth_identity SET type = 'login_id' WHERE type = 'password';
