-- Put upgrade SQL here
ALTER TABLE _auth_provider_custom_token ADD COLUMN raw_profile JSONB;
UPDATE _auth_provider_custom_token SET raw_profile = '{}';
ALTER TABLE _auth_provider_custom_token ALTER COLUMN raw_profile SET NOT NULL;
