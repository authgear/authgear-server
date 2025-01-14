-- +migrate Up
ALTER TABLE _auth_identity_login_id ADD COLUMN login_id_type TEXT NOT NULL;

-- +migrate Down
ALTER TABLE _auth_identity_login_id DROP COLUMN login_id_type;
