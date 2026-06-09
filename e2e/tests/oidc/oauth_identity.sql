{{ $oauth_identity_id := uuidv4 }}

INSERT INTO _auth_identity (
  id, app_id, type, user_id, created_at, updated_at
)
SELECT
  '{{ $oauth_identity_id }}',
  '{{ .AppID }}',
  'oauth',
  i.user_id,
  NOW(),
  NOW()
FROM _auth_identity i
JOIN _auth_identity_login_id l ON i.id = l.id AND l.app_id = i.app_id
WHERE i.app_id = '{{ .AppID }}'
AND l.login_id = 'e2e_userinfo_user@example.com';

INSERT INTO _auth_identity_oauth (
  id, app_id, provider_type, provider_keys, provider_user_id, claims, profile
)
VALUES (
  '{{ $oauth_identity_id }}',
  '{{ .AppID }}',
  'google',
  '{}',
  'google-subject-123',
  '{}',
  '{}'
);
