-- Put upgrade SQL here
ALTER TABLE _auth_provider_password ADD COLUMN claims JSONB;
UPDATE _auth_provider_password SET claims = '{}'::JSONB;
ALTER TABLE _auth_provider_password ALTER COLUMN claims SET NOT NULL;

ALTER TABLE _auth_provider_oauth ADD COLUMN claims JSONB;
UPDATE _auth_provider_oauth SET claims = '{}'::JSONB;
ALTER TABLE _auth_provider_oauth ALTER COLUMN claims SET NOT NULL;

ALTER TABLE _auth_provider_custom_token ADD COLUMN claims JSONB;
UPDATE _auth_provider_custom_token SET claims = '{}'::JSONB;
ALTER TABLE _auth_provider_custom_token ALTER COLUMN claims SET NOT NULL;
