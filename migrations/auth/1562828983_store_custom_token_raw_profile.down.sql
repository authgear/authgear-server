-- Put downgrade SQL here
ALTER TABLE _auth_provider_custom_token DROP COLUMN raw_profile;
