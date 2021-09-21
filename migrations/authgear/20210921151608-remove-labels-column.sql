-- +migrate Up
ALTER TABLE _auth_authenticator DROP COLUMN labels;
ALTER TABLE _auth_identity DROP COLUMN labels;
ALTER TABLE _auth_oauth_authorization DROP COLUMN labels;
ALTER TABLE _auth_user DROP COLUMN labels;

-- +migrate Down
ALTER TABLE _auth_authenticator ADD COLUMN labels jsonb;
UPDATE _auth_authenticator SET labels = '{}'::jsonb;
ALTER TABLE _auth_authenticator ALTER COLUMN labels SET NOT NULL;

ALTER TABLE _auth_identity ADD COLUMN labels jsonb;
UPDATE _auth_identity SET labels = '{}'::jsonb;
ALTER TABLE _auth_identity ALTER COLUMN labels SET NOT NULL;

ALTER TABLE _auth_oauth_authorization ADD COLUMN labels jsonb;
UPDATE _auth_oauth_authorization SET labels = '{}'::jsonb;
ALTER TABLE _auth_oauth_authorization ALTER COLUMN labels SET NOT NULL;

ALTER TABLE _auth_user ADD COLUMN labels jsonb;
UPDATE _auth_user SET labels = '{}'::jsonb;
ALTER TABLE _auth_user ALTER COLUMN labels SET NOT NULL;
