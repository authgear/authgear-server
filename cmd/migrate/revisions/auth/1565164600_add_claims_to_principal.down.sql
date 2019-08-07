-- Put downgrade SQL here
ALTER TABLE _auth_provider_password DROP COLUMN claims;
ALTER TABLE _auth_provider_oauth DROP COLUMN claims;
ALTER TABLE _auth_provider_custom_token DROP COLUMN claims;
