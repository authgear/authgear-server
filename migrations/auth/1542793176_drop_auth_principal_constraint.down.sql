ALTER TABLE _auth_principal ADD CONSTRAINT _auth_principal_user_id_provider_key UNIQUE (user_id, provider);
