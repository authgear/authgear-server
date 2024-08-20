{{ $userID := (uuidv4) }}
{{ $identityID := (uuidv4) }}

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
    '{{ $userID }}',
    '{{ .AppID }}',
    NOW(),
    NOW(),
    NULL,
    NOW(),
    'f',
    NULL,
    '{"email": "jdoe@example.com"}',
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
    '{{ $identityID }}',
    '{{ .AppID }}',
    'ldap',
    '{{ $userID }}',
    NOW(),
    NOW()
  );

INSERT INTO
  _auth_identity_ldap (
    "id",
    "app_id",
    "server_name",
    "user_id_attribute_name",
    "user_id_attribute_value",
    "claims",
    "raw_entry_json"
  )
VALUES
  (
    '{{ $identityID }}',
    '{{ .AppID }}',
    'ldap-server-1',
    'uid',
    'jdoe',
    '{"email": "jdoe@example.com"}',
    '{"dn": "cn=jdoe,ou=people,ou=HK,dc=authgear,dc=com"}'
  );
