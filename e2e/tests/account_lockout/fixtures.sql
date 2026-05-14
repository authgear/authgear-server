-- Create test user for account lockout tests
INSERT INTO _auth_user (
  "id",
  "app_id",
  "created_at",
  "updated_at",
  "is_disabled",
  "disable_reason",
  "standard_attributes",
  "custom_attributes",
  "is_deactivated",
  "is_anonymized"
) VALUES (
  '{{ uuidv4 }}',
  '{{ .AppID }}',
  NOW(),
  NOW(),
  false,
  NULL,
  '{"email": "lockout_test_{{ .AppID }}@example.com"}',
  '{}',
  false,
  false
);

INSERT INTO _auth_identity (
  "id",
  "app_id",
  "type",
  "user_id",
  "created_at",
  "updated_at"
) VALUES (
  '{{ uuidv4 }}',
  '{{ .AppID }}',
  'login_id',
  (SELECT "id" FROM "_auth_user" WHERE "app_id" = '{{ .AppID }}' ORDER BY "created_at" DESC LIMIT 1),
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
  (SELECT "id" FROM "_auth_identity" WHERE "app_id" = '{{ .AppID }}' AND "type" = 'login_id' ORDER BY "created_at" DESC LIMIT 1),
  '{{ .AppID }}',
  'email',
  'lockout_test_{{ .AppID }}@example.com',
  '{"email": "lockout_test_{{ .AppID }}@example.com"}',
  'lockout_test_{{ .AppID }}@example.com',
  'lockout_test_{{ .AppID }}@example.com',
  'email'
);
