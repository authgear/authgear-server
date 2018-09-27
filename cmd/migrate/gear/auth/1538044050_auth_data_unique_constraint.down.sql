BEGIN;

ALTER TABLE _auth_provider_password DROP CONSTRAINT "_auth_provider_password_auth_data_key";

END;
