{{ $userID := (uuidv4) }}
{{ $emailID := (uuidv4) }}
{{ $phoneID := (uuidv4) }}
{{ $passwordID := (uuidv4) }}

INSERT INTO _auth_user (
  "id",
  "app_id",
  "created_at",
  "updated_at",
  "is_disabled",
  "standard_attributes"
) VALUES (
  '{{ $userID }}',
  '{{ .AppID }}',
  NOW(),
  NOW(),
  FALSE,
  '{"email": "user@example.com", "phone_number": "+8525100000" }'
);

INSERT INTO _auth_identity (
  "id",
  "app_id",
  "type",
  "user_id",
  "created_at",
  "updated_at"
) VALUES (
  '{{ $phoneID }}',
  '{{ .AppID }}',
  'login_id',
  '{{ $userID }}',
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
  '{{ $phoneID }}',
  '{{ .AppID }}',
  'phone',
  '+85251000000',
  '{"phone_number": "+85251000000"}',
  '+85251000000',
  '+85251000000',
  'phone'
);

INSERT INTO _auth_identity (
  "id",
  "app_id",
  "type",
  "user_id",
  "created_at",
  "updated_at"
) VALUES (
  '{{ $emailID }}',
  '{{ .AppID }}',
  'login_id',
  '{{ $userID }}',
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
  '{{ $emailID }}',
  '{{ .AppID }}',
  'email',
  'user@example.com',
  '{"email": "user@example.com"}',
  'user@example.com',
  'user@example.com',
  'email'
);

INSERT INTO _auth_authenticator (
  "id",
  "app_id",
  "type",
  "user_id",
  "created_at",
  "updated_at",
  "is_default",
  "kind"
) VALUES (
  '{{ $passwordID }}',
  '{{ .AppID }}',
  'password',
  '{{ $userID }}',
  NOW(),
  NOW(),
  FALSE,
  'primary'
);

INSERT INTO _auth_authenticator_password (
  "id",
  "app_id",
  "password_hash"
) VALUES (
  '{{ $passwordID }}',
  '{{ .AppID }}',
  '$bcrypt-sha512$$2a$10$x/wTJRgyuj4J1Pi/Kv7TYOd8kqV/hsU4nHkQ4.y.CVUz4cGXeHI7q'
);
