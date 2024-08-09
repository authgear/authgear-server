{{ $user_id := (uuidv4) }}
{{ $identity_id := (uuidv4) }}
{{ $authenticator_password_id := (uuidv4) }}
{{ $authenticator_totp_id := (uuidv4) }}
{{ $password_history_id := (uuidv4) }}
{{ $recovery_code_id_1 := (uuidv4) }}
{{ $recovery_code_id_2 := (uuidv4) }}
{{ $recovery_code_id_3 := (uuidv4) }}
{{ $recovery_code_id_4 := (uuidv4) }}
{{ $recovery_code_id_5 := (uuidv4) }}
{{ $recovery_code_id_6 := (uuidv4) }}
{{ $recovery_code_id_7 := (uuidv4) }}
{{ $recovery_code_id_8 := (uuidv4) }}
{{ $recovery_code_id_9 := (uuidv4) }}
{{ $recovery_code_id_10 := (uuidv4) }}
{{ $recovery_code_id_11 := (uuidv4) }}
{{ $recovery_code_id_12 := (uuidv4) }}
{{ $recovery_code_id_13 := (uuidv4) }}
{{ $recovery_code_id_14 := (uuidv4) }}
{{ $recovery_code_id_15 := (uuidv4) }}
{{ $recovery_code_id_16 := (uuidv4) }}
INSERT INTO
  _auth_user (
    "id",
    "app_id",
    "created_at",
    "updated_at",
    "last_login_at",
    "login_at",
    "is_disabled",
    "disable_reason",
    "standard_attributes",
    "custom_attributes",
    "is_deactivated",
    "delete_at",
    "is_anonymized",
    "anonymize_at",
    "anonymized_at",
    "last_indexed_at",
    "require_reindex_after"
  )
VALUES
  (
    '{{ $user_id }}',
    '{{ .AppID }}',
    NOW(),
    NOW(),
    NULL,
    NOW(),
    'f',
    NULL,
    '{"email": "signup@example.com"}',
    '{}',
    'f',
    NULL,
    'f',
    NULL,
    NULL,
    NOW(),
    NOW()
  );
INSERT INTO
  _auth_identity (
    "id",
    "app_id",
    "type",
    "user_id",
    "created_at",
    "updated_at"
  )
VALUES
  (
    '{{ $identity_id }}',
    '{{ .AppID }}',
    'login_id',
    '{{ $user_id }}',
    NOW(),
    NOW()
  );
INSERT INTO
  _auth_authenticator (
    "id",
    "app_id",
    "type",
    "user_id",
    "created_at",
    "updated_at",
    "is_default",
    "kind"
  )
VALUES
  (
    '{{ $authenticator_password_id }}',
    '{{ .AppID }}',
    'password',
    '{{ $user_id }}',
    NOW(),
    NOW(),
    't',
    'primary'
  ),
  (
    '{{ $authenticator_totp_id }}',
    '{{ .AppID }}',
    'totp',
    '{{ $user_id }}',
    NOW(),
    NOW(),
    't',
    'secondary'
  );
INSERT INTO
  _auth_authenticator_password ("id", "app_id", "password_hash", "expire_after")
VALUES
  (
    '{{ $authenticator_password_id }}',
    '{{ .AppID }}',
    '$bcrypt-sha512$$2a$10$TsJ6RYa.EL46SbDLGQnwTeFYi.3gdBiPWtO.J05zo0zm1yuNO6/6K',
    NULL
  );
INSERT INTO
  _auth_authenticator_totp ("id", "app_id", "secret", "display_name")
VALUES
  (
    '{{ $authenticator_totp_id }}',
    '{{ .AppID }}',
    '3I526Y3Y7GSXO34RTFEEFXCJM6Y4VZXR',
    'TOTP @ 2024-06-26T15:11:48Z'
  );
INSERT INTO
  _auth_identity_login_id (
    "id",
    "app_id",
    "login_id_key",
    "login_id",
    "claims",
    "original_login_id",
    "unique_key",
    "login_id_type"
  )
VALUES
  (
    '{{ $identity_id }}',
    '{{ .AppID }}',
    'email',
    'signup@example.com',
    '{"email": "signup@example.com"}',
    'signup@example.com',
    'signup@example.com',
    'email'
  );
INSERT INTO
  _auth_password_history (
    "id",
    "app_id",
    "created_at",
    "user_id",
    "password"
  )
VALUES
  (
    '{{ $password_history_id }}',
    '{{ .AppID }}',
    NOW(),
    '{{ $user_id }}',
    '$bcrypt-sha512$$2a$10$TsJ6RYa.EL46SbDLGQnwTeFYi.3gdBiPWtO.J05zo0zm1yuNO6/6K'
  );
INSERT INTO
  _auth_recovery_code (
    "id",
    "app_id",
    "user_id",
    "code",
    "created_at",
    "consumed",
    "updated_at"
  )
VALUES
  (
    '{{ $recovery_code_id_1 }}',
    '{{ .AppID }}',
    '{{ $user_id }}',
    'DZ9EDP179S',
    NOW(),
    'f',
    NOW()
  ),
  (
    '{{ $recovery_code_id_2 }}',
    '{{ .AppID }}',
    '{{ $user_id }}',
    'Q5C26V77PF',
    NOW(),
    'f',
    NOW()
  ),
  (
    '{{ $recovery_code_id_3 }}',
    '{{ .AppID }}',
    '{{ $user_id }}',
    'FMG6JAJEYR',
    NOW(),
    'f',
    NOW()
  ),
  (
    '{{ $recovery_code_id_4 }}',
    '{{ .AppID }}',
    '{{ $user_id }}',
    'HJZQHR527J',
    NOW(),
    'f',
    NOW()
  ),
  (
    '{{ $recovery_code_id_5 }}',
    '{{ .AppID }}',
    '{{ $user_id }}',
    'BVK0KM646A',
    NOW(),
    'f',
    NOW()
  ),
  (
    '{{ $recovery_code_id_6 }}',
    '{{ .AppID }}',
    '{{ $user_id }}',
    '64F4SDC0F8',
    NOW(),
    'f',
    NOW()
  ),
  (
    '{{ $recovery_code_id_7 }}',
    '{{ .AppID }}',
    '{{ $user_id }}',
    'NCCJSPC5Q5',
    NOW(),
    'f',
    NOW()
  ),
  (
    '{{ $recovery_code_id_8 }}',
    '{{ .AppID }}',
    '{{ $user_id }}',
    'T5VBWZP3VH',
    NOW(),
    'f',
    NOW()
  ),
  (
    '{{ $recovery_code_id_9 }}',
    '{{ .AppID }}',
    '{{ $user_id }}',
    '3Y3CDKKHZ6',
    NOW(),
    'f',
    NOW()
  ),
  (
    '{{ $recovery_code_id_10 }}',
    '{{ .AppID }}',
    '{{ $user_id }}',
    'E2E5WT3XY8',
    NOW(),
    'f',
    NOW()
  ),
  (
    '{{ $recovery_code_id_11 }}',
    '{{ .AppID }}',
    '{{ $user_id }}',
    'P6MFB1M4SQ',
    NOW(),
    'f',
    NOW()
  ),
  (
    '{{ $recovery_code_id_12 }}',
    '{{ .AppID }}',
    '{{ $user_id }}',
    '2R15Y5TE4Q',
    NOW(),
    'f',
    NOW()
  ),
  (
    '{{ $recovery_code_id_13 }}',
    '{{ .AppID }}',
    '{{ $user_id }}',
    '0CG2SA9JA0',
    NOW(),
    'f',
    NOW()
  ),
  (
    '{{ $recovery_code_id_14 }}',
    '{{ .AppID }}',
    '{{ $user_id }}',
    '9HQCS8HTHT',
    NOW(),
    'f',
    NOW()
  ),
  (
    '{{ $recovery_code_id_15 }}',
    '{{ .AppID }}',
    '{{ $user_id }}',
    'G25RG95RZ7',
    NOW(),
    'f',
    NOW()
  ),
  (
    '{{ $recovery_code_id_16 }}',
    '{{ .AppID }}',
    '{{ $user_id }}',
    'XGG202XTDZ',
    NOW(),
    'f',
    NOW()
  );
