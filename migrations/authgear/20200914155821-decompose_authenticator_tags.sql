-- +migrate Up

ALTER TABLE _auth_authenticator ADD COLUMN is_default boolean;
UPDATE _auth_authenticator SET is_default = tag ? 'authentication:default_authenticator';
ALTER TABLE _auth_authenticator ALTER COLUMN is_default SET NOT NULL;

ALTER TABLE _auth_authenticator ADD COLUMN kind text;
UPDATE _auth_authenticator SET kind =
    CASE
        WHEN tag ? 'authentication:primary_authenticator' THEN 'primary'
        WHEN tag ? 'authentication:secondary_authenticator' THEN 'secondary'
    END;
ALTER TABLE _auth_authenticator ALTER COLUMN kind SET NOT NULL;

ALTER TABLE _auth_authenticator DROP COLUMN tag;

-- +migrate Down

ALTER TABLE _auth_authenticator ADD COLUMN tag JSONB;
UPDATE _auth_authenticator SET tag = '[]'::jsonb;
ALTER TABLE _auth_authenticator ALTER COLUMN tag SET NOT NULL;

ALTER TABLE _auth_authenticator DROP COLUMN is_default;
ALTER TABLE _auth_authenticator DROP COLUMN kind;
