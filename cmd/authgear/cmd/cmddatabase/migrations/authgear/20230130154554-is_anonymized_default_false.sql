-- +migrate Up

ALTER TABLE _auth_user ALTER COLUMN is_anonymized SET DEFAULT FALSE;

-- +migrate Down

ALTER TABLE _auth_user ALTER COLUMN is_anonymized DROP DEFAULT;