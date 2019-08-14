ALTER TABLE _auth_provider_password ALTER COLUMN login_id TYPE jsonb USING login_id::text::jsonb;
ALTER TABLE _auth_provider_password RENAME COLUMN login_id TO auth_data;
ALTER TABLE _auth_provider_password DROP login_id_key;
