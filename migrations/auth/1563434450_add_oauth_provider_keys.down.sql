-- Put downgrade SQL here
ALTER TABLE _auth_provider_oauth DROP CONSTRAINT _auth_provider_oauth_provider_user_id;

ALTER TABLE _auth_provider_oauth DROP COLUMN provider_keys;

ALTER TABLE _auth_provider_oauth RENAME COLUMN provider_type to oauth_provider;

ALTER TABLE _auth_provider_oauth ADD CONSTRAINT _auth_provider_oauth_provider_user_id UNIQUE (oauth_provider, provider_user_id);
