ALTER TABLE _auth_provider_password ALTER COLUMN auth_data TYPE jsonb USING auth_data::text::jsonb;
ALTER TABLE _auth_provider_password DROP auth_data_key;
