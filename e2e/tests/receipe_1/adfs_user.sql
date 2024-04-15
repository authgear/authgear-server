INSERT INTO _auth_identity_oauth ("id", "app_id", "provider_type", "provider_keys", "provider_user_id", "claims", "profile")
VALUES (
  (SELECT id FROM _auth_identity_login_id WHERE login_id = 'mock@example.com' AND app_id = '{{ .AppID }}' LIMIT 1),
  '{{ .AppID }}',
  'adfs',
  '{}',
  'mock',
  '{"email": "mock@example.com"}',
  '{"email": "mock@example.com"}'
);
