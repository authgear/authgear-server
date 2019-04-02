ALTER TABLE _auth_provider_password ALTER COLUMN auth_data TYPE text;
ALTER TABLE _auth_provider_password ADD auth_data_key text;