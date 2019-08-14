ALTER TABLE _auth_provider_oauth ADD CONSTRAINT "_auth_provider_oauth_provider_user_id" UNIQUE ("oauth_provider", "provider_user_id");
