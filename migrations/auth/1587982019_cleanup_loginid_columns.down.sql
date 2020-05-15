ALTER TABLE _auth_identity_login_id ADD COLUMN password TEXT;
ALTER TABLE _auth_identity_login_id ADD COLUMN realm TEXT;
UPDATE _auth_identity_login_id SET password = '', realm = '';
ALTER TABLE _auth_identity_login_id ALTER COLUMN password SET NOT NULL;
ALTER TABLE _auth_identity_login_id ALTER COLUMN realm SET NOT NULL;
UPDATE _auth_identity SET type = 'password' WHERE type = 'login_id';
