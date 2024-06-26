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
    'cb3a2f82-cbed-471b-aa42-0bf8da2b74cd',
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
  ) ON CONFLICT DO NOTHING;
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
    '0b082939-1c83-4c23-930f-8b3072beb1ca',
    '{{ .AppID }}',
    'login_id',
    'cb3a2f82-cbed-471b-aa42-0bf8da2b74cd',
    NOW(),
    NOW()
  ) ON CONFLICT DO NOTHING;
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
    '248cbeb9-1aa8-4346-84b3-ec1b6e6030cd',
    '{{ .AppID }}',
    'password',
    'cb3a2f82-cbed-471b-aa42-0bf8da2b74cd',
    NOW(),
    NOW(),
    't',
    'primary'
  ),
  (
    '58495140-8f35-4759-b24b-e219783ffb0d',
    '{{ .AppID }}',
    'totp',
    'cb3a2f82-cbed-471b-aa42-0bf8da2b74cd',
    NOW(),
    NOW(),
    't',
    'secondary'
  ) ON CONFLICT DO NOTHING;
INSERT INTO
  _auth_authenticator_password ("id", "app_id", "password_hash", "expire_after")
VALUES
  (
    '248cbeb9-1aa8-4346-84b3-ec1b6e6030cd',
    '{{ .AppID }}',
    '$bcrypt-sha512$$2a$10$TsJ6RYa.EL46SbDLGQnwTeFYi.3gdBiPWtO.J05zo0zm1yuNO6/6K',
    NULL
  ) ON CONFLICT DO NOTHING;
INSERT INTO
  _auth_authenticator_totp ("id", "app_id", "secret", "display_name")
VALUES
  (
    '58495140-8f35-4759-b24b-e219783ffb0d',
    '{{ .AppID }}',
    '3I526Y3Y7GSXO34RTFEEFXCJM6Y4VZXR',
    'TOTP @ 2024-06-26T15:11:48Z'
  ) ON CONFLICT DO NOTHING;
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
    '0b082939-1c83-4c23-930f-8b3072beb1ca',
    '{{ .AppID }}',
    'email',
    'signup@example.com',
    '{"email": "signup@example.com"}',
    'signup@example.com',
    'signup@example.com',
    'email'
  ) ON CONFLICT DO NOTHING;
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
    'b5de2ef7-f733-4dde-80fd-8212061ebcdd',
    '{{ .AppID }}',
    NOW(),
    'cb3a2f82-cbed-471b-aa42-0bf8da2b74cd',
    '$bcrypt-sha512$$2a$10$TsJ6RYa.EL46SbDLGQnwTeFYi.3gdBiPWtO.J05zo0zm1yuNO6/6K'
  ) ON CONFLICT DO NOTHING;
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
    '107f9680-7cc3-4a43-856c-d02ce13dd8fd',
    '{{ .AppID }}',
    'cb3a2f82-cbed-471b-aa42-0bf8da2b74cd',
    'DZ9EDP179S',
    NOW(),
    'f',
    NOW()
  ),
  (
    '2bec8b35-db73-42ef-88c4-e7a7b892513d',
    '{{ .AppID }}',
    'cb3a2f82-cbed-471b-aa42-0bf8da2b74cd',
    'Q5C26V77PF',
    NOW(),
    'f',
    NOW()
  ),
  (
    '3ab62f59-0396-4292-9022-82d37734ffbd',
    '{{ .AppID }}',
    'cb3a2f82-cbed-471b-aa42-0bf8da2b74cd',
    'FMG6JAJEYR',
    NOW(),
    'f',
    NOW()
  ),
  (
    '4387261b-c761-4c33-be12-b932683fb8cd',
    '{{ .AppID }}',
    'cb3a2f82-cbed-471b-aa42-0bf8da2b74cd',
    'HJZQHR527J',
    NOW(),
    'f',
    NOW()
  ),
  (
    '5b9d267f-ec51-4107-8d5f-63069681284d',
    '{{ .AppID }}',
    'cb3a2f82-cbed-471b-aa42-0bf8da2b74cd',
    'BVK0KM646A',
    NOW(),
    'f',
    NOW()
  ),
  (
    '5f9bf330-bca0-488b-b48b-02b23630583d',
    '{{ .AppID }}',
    'cb3a2f82-cbed-471b-aa42-0bf8da2b74cd',
    '64F4SDC0F8',
    NOW(),
    'f',
    NOW()
  ),
  (
    '6fba64a0-56b6-4100-a210-14278784f93d',
    '{{ .AppID }}',
    'cb3a2f82-cbed-471b-aa42-0bf8da2b74cd',
    'NCCJSPC5Q5',
    NOW(),
    'f',
    NOW()
  ),
  (
    '7a5b0c9e-df46-4bf5-97db-59313eff0c7d',
    '{{ .AppID }}',
    'cb3a2f82-cbed-471b-aa42-0bf8da2b74cd',
    'T5VBWZP3VH',
    NOW(),
    'f',
    NOW()
  ),
  (
    '8676e790-7e34-4cd9-b548-58fb54386bcd',
    '{{ .AppID }}',
    'cb3a2f82-cbed-471b-aa42-0bf8da2b74cd',
    '3Y3CDKKHZ6',
    NOW(),
    'f',
    NOW()
  ),
  (
    '9efa37a1-9d75-4400-8aea-400afd7c84fd',
    '{{ .AppID }}',
    'cb3a2f82-cbed-471b-aa42-0bf8da2b74cd',
    'E2E5WT3XY8',
    NOW(),
    'f',
    NOW()
  ),
  (
    'aeb1d7b9-93ae-4e75-9cf8-e823b916993d',
    '{{ .AppID }}',
    'cb3a2f82-cbed-471b-aa42-0bf8da2b74cd',
    'P6MFB1M4SQ',
    NOW(),
    'f',
    NOW()
  ),
  (
    'bc44c7ad-2ab1-44a6-807e-ce304a2de89d',
    '{{ .AppID }}',
    'cb3a2f82-cbed-471b-aa42-0bf8da2b74cd',
    '2R15Y5TE4Q',
    NOW(),
    'f',
    NOW()
  ),
  (
    'bd820c3a-33a0-4708-bcdf-5b4826417aed',
    '{{ .AppID }}',
    'cb3a2f82-cbed-471b-aa42-0bf8da2b74cd',
    '0CG2SA9JA0',
    NOW(),
    'f',
    NOW()
  ),
  (
    'bdd4713b-d301-4b67-881e-7bccd0bf426d',
    '{{ .AppID }}',
    'cb3a2f82-cbed-471b-aa42-0bf8da2b74cd',
    '9HQCS8HTHT',
    NOW(),
    'f',
    NOW()
  ),
  (
    'c89b834f-82f1-4da6-843f-93aaea7d121d',
    '{{ .AppID }}',
    'cb3a2f82-cbed-471b-aa42-0bf8da2b74cd',
    'G25RG95RZ7',
    NOW(),
    'f',
    NOW()
  ),
  (
    'd0949d1b-2f27-4196-ad4d-a9335d72288d',
    '{{ .AppID }}',
    'cb3a2f82-cbed-471b-aa42-0bf8da2b74cd',
    'XGG202XTDZ',
    NOW(),
    'f',
    NOW()
  ) ON CONFLICT DO NOTHING;