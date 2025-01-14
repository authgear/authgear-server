-- +migrate Up

ALTER TABLE _auth_user ADD COLUMN is_anonymized boolean;
UPDATE _auth_user SET is_anonymized = FALSE;
ALTER TABLE _auth_user ALTER COLUMN is_anonymized SET NOT NULL;

ALTER TABLE _auth_user ADD COLUMN anonymize_at timestamp without time zone;
CREATE INDEX _auth_user_anonymize_at ON _auth_user USING BRIN (anonymize_at);

ALTER TABLE _auth_user ADD COLUMN anonymized_at timestamp without time zone;
CREATE INDEX _auth_user_anonymized_at ON _auth_user USING BRIN (anonymized_at);

-- +migrate Down

ALTER TABLE _auth_user DROP COLUMN is_anonymized;
ALTER TABLE _auth_user DROP COLUMN anonymize_at;
ALTER TABLE _auth_user DROP COLUMN anonymized_at;