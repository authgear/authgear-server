INSERT INTO _auth_identity (
  "id",
  "app_id",
  "type",
  "user_id",
  "created_at",
  "updated_at"
) VALUES (
  '{{ .AppID }}__google__',
  '{{ .AppID }}',
  'oauth',
  (SELECT b.user_id
    FROM _auth_identity_login_id a
    JOIN _auth_identity b
    ON a.id = b.id
    WHERE a.login_id = 'mock'
    AND a.app_id = '{{ .AppID }}'
    LIMIT 1
  ),
  NOW(),
  NOW()
);

INSERT INTO _auth_identity_oauth (
  "id",
  "app_id",
  "provider_type",
  "provider_keys",
  "provider_user_id",
  "claims",
  "profile"
) VALUES (
  '{{ .AppID }}__google__',
  '{{ .AppID }}',
  'google',
  '{}',
  'mock',
  '{"email": "mock@example.com"}',
  '{"email": "mock@example.com"}'
);
