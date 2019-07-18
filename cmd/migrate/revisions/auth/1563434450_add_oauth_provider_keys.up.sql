-- Put upgrade SQL here
ALTER TABLE _auth_provider_oauth DROP CONSTRAINT _auth_provider_oauth_provider_user_id;

ALTER TABLE _auth_provider_oauth ADD COLUMN provider_keys JSONB;
ALTER TABLE _auth_provider_oauth ALTER COLUMN provider_keys SET DEFAULT '{}'::JSONB;
UPDATE _auth_provider_oauth SET provider_keys = '{}'::JSONB;
ALTER TABLE _auth_provider_oauth ALTER COLUMN provider_keys SET NOT NULL;

ALTER TABLE _auth_provider_oauth RENAME COLUMN oauth_provider to provider_type;

ALTER TABLE _auth_provider_oauth ADD CONSTRAINT _auth_provider_oauth_provider_user_id UNIQUE (provider_type, provider_keys, provider_user_id);
