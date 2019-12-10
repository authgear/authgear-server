ALTER TABLE _auth_provider_password ADD COLUMN original_login_id TEXT;
UPDATE _auth_provider_password SET original_login_id=login_id;
ALTER TABLE _auth_provider_password ALTER COLUMN original_login_id SET NOT NULL;

ALTER TABLE _auth_provider_password ADD COLUMN unique_key TEXT;
UPDATE _auth_provider_password SET unique_key=lower(login_id);
ALTER TABLE _auth_provider_password ALTER COLUMN unique_key SET NOT NULL;
ALTER TABLE _auth_provider_password ADD CONSTRAINT _auth_provider_password_unique_key_realm UNIQUE(app_id, unique_key, realm);

UPDATE _auth_provider_password SET login_id=lower(login_id);
UPDATE _auth_provider_password SET claims = jsonb_set(claims, '{email}', concat('"', lower(claims->>'email'), '"')::jsonb, false);
UPDATE _auth_provider_password SET claims = jsonb_set(claims, '{username}', concat('"', lower(claims->>'username'), '"')::jsonb, false);
