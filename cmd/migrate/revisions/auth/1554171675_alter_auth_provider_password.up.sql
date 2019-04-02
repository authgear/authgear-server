ALTER TABLE _auth_provider_password ALTER COLUMN auth_data TYPE text;
ALTER TABLE _auth_provider_password RENAME COLUMN auth_data TO login_id;
ALTER TABLE _auth_provider_password ADD login_id_key text;