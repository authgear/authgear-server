{{ $user_id := (uuidv4) }}
{{ $email_id := (uuidv4) }}

INSERT INTO _auth_user (
  "id",
  "app_id",
  "created_at",
  "updated_at",
  "is_disabled",
  "standard_attributes"
) VALUES (
  '{{ $user_id }}',
  '{{ .AppID }}',
  NOW(),
  NOW(),
  FALSE,
  '{"email": "recipient@example.com"}'
);

INSERT INTO _auth_identity (
  "id",
  "app_id",
  "type",
  "user_id",
  "created_at",
  "updated_at"
) VALUES (
  '{{ $email_id }}',
  '{{ .AppID }}',
  'login_id',
  '{{ $user_id }}',
  NOW(),
  NOW()
);

INSERT INTO _auth_identity_login_id (
  "id",
  "app_id",
  "login_id_key",
  "login_id",
  "claims",
  "original_login_id",
  "unique_key",
  "login_id_type"
) VALUES (
  '{{ $email_id }}',
  '{{ .AppID }}',
  'email',
  'recipient@example.com',
  '{"email": "recipient@example.com"}',
  'recipient@example.com',
  'recipient@example.com',
  'email'
);
