{{ $user_id := (uuidv4) }}
{{ $phone_id := (uuidv4) }}
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
  '{"email": "user@example.com", "phone_number": "+85251000000"}'
);

INSERT INTO _auth_identity (
  "id",
  "app_id",
  "type",
  "user_id",
  "created_at",
  "updated_at"
) VALUES (
  '{{ $phone_id }}',
  '{{ .AppID }}',
  'login_id',
  '{{ $user_id }}',
  NOW(),
  NOW()
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
  '{{ $phone_id }}',
  '{{ .AppID }}',
  'phone',
  '+85251000000',
  '{"phone_number": "+85251000000"}',
  '+85251000000',
  '+85251000000',
  'phone'
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
  'user@example.com',
  '{"email": "user@example.com"}',
  'user@example.com',
  'user@example.com',
  'email'
);
