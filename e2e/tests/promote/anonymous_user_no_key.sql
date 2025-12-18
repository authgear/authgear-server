{{ $userID := (uuidv4) }}
{{ $identityID := (uuidv4) }}

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
  '{"testuserid": "1"}'
);

INSERT INTO _auth_identity (
  "id",
  "app_id",
  "type",
  "user_id",
  "created_at",
  "updated_at"
) VALUES (
  '{{ $identityID }}',
  '{{ .AppID }}',
  'anonymous',
  '{{ $userID }}',
  NOW(),
  NOW()
);

INSERT INTO _auth_identity_anonymous (
  "id",
  "app_id",
  "key_id",
  "key"
) VALUES (
  '{{ $identityID }}',
  '{{ .AppID }}',
  '',
  NULL
);
