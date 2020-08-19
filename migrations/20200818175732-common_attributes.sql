-- +migrate Up
ALTER TABLE _auth_recovery_code ADD COLUMN updated_at timestamp without time zone;
UPDATE _auth_recovery_code SET updated_at = created_at;
ALTER TABLE _auth_recovery_code ALTER COLUMN updated_at SET NOT NULL;

ALTER TABLE _auth_identity ADD COLUMN created_at timestamp without time zone;
ALTER TABLE _auth_identity ADD COLUMN updated_at timestamp without time zone;
UPDATE _auth_identity i SET created_at = NOW(), updated_at = NOW()
    FROM _auth_identity_login_id il WHERE i.id = il.id;
UPDATE _auth_identity i SET created_at = NOW(), updated_at = NOW()
    FROM _auth_identity_anonymous ia WHERE i.id = ia.id;
UPDATE _auth_identity i SET created_at = io.created_at, updated_at = io.updated_at
    FROM _auth_identity_oauth io WHERE i.id = io.id;
ALTER TABLE _auth_identity ALTER COLUMN created_at SET NOT NULL;
ALTER TABLE _auth_identity ALTER COLUMN updated_at SET NOT NULL;
ALTER TABLE _auth_identity_oauth DROP COLUMN created_at;
ALTER TABLE _auth_identity_oauth DROP COLUMN updated_at;

ALTER TABLE _auth_authenticator ADD COLUMN created_at timestamp without time zone;
ALTER TABLE _auth_authenticator ADD COLUMN updated_at timestamp without time zone;
UPDATE _auth_authenticator a SET created_at = ao.created_at, updated_at = ao.created_at
FROM _auth_authenticator_oob ao WHERE a.id = ao.id;
UPDATE _auth_authenticator a SET created_at = at.created_at, updated_at = at.created_at
FROM _auth_authenticator_totp at WHERE a.id = at.id;
UPDATE _auth_authenticator a SET created_at = NOW(), updated_at = NOW()
FROM _auth_authenticator_password ap WHERE a.id = ap.id;
ALTER TABLE _auth_authenticator ALTER COLUMN created_at SET NOT NULL;
ALTER TABLE _auth_authenticator ALTER COLUMN updated_at SET NOT NULL;
ALTER TABLE _auth_authenticator_oob DROP COLUMN created_at;
ALTER TABLE _auth_authenticator_totp DROP COLUMN created_at;

-- +migrate Down
ALTER TABLE _auth_recovery_code DROP COLUMN updated_at;

ALTER TABLE _auth_identity DROP COLUMN created_at;
ALTER TABLE _auth_identity DROP COLUMN updated_at;
ALTER TABLE _auth_identity_oauth ADD COLUMN created_at timestamp without time zone;
ALTER TABLE _auth_identity_oauth ADD COLUMN updated_at timestamp without time zone;
UPDATE _auth_identity_oauth SET created_at = NOW(), updated_at = NOW();
ALTER TABLE _auth_identity_oauth ALTER COLUMN created_at SET NOT NULL;
ALTER TABLE _auth_identity_oauth ALTER COLUMN updated_at SET NOT NULL;

ALTER TABLE _auth_authenticator DROP COLUMN created_at;
ALTER TABLE _auth_authenticator DROP COLUMN updated_at;
ALTER TABLE _auth_authenticator_oob ADD COLUMN created_at timestamp without time zone;
ALTER TABLE _auth_authenticator_totp ADD COLUMN created_at timestamp without time zone;
UPDATE _auth_authenticator_oob SET created_at = NOW();
UPDATE _auth_authenticator_totp SET created_at = NOW();
ALTER TABLE _auth_authenticator_oob ALTER COLUMN created_at SET NOT NULL;
ALTER TABLE _auth_authenticator_totp ALTER COLUMN created_at SET NOT NULL;
