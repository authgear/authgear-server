{{ $user_id := (uuidv4) }}
{{ $identity_id := (uuidv4) }}

INSERT INTO _auth_user
  ("id", "app_id", "created_at", "updated_at", "last_login_at", "login_at", "is_disabled", "disable_reason", "standard_attributes", "custom_attributes", "is_deactivated", "delete_at", "is_anonymized", "anonymize_at", "anonymized_at", "last_indexed_at", "require_reindex_after")
VALUES
  ('{{ $user_id }}', '{{ .AppID }}', '2024-08-27 07:51:42.040683', '2024-08-27 07:51:42.056654', NULL, NULL, 'f', NULL, '{"preferred_username": "e2e_stored", "given_name": "A"}', '{}', 'f', NULL, 'f', NULL, NULL, '2024-08-27 07:51:42.079532', '2024-08-27 07:51:42.059056');

INSERT INTO _auth_identity
  ("id", "app_id", "type", "user_id", "created_at", "updated_at")
VALUES
  ('{{ $identity_id }}', '{{ .AppID }}', 'login_id', '{{ $user_id }}', '2024-08-27 07:51:42.051107', '2024-08-27 07:51:42.051107');

INSERT INTO _auth_identity_login_id
  ("id", "app_id", "login_id_key", "login_id", "claims", "original_login_id", "unique_key", "login_id_type")
VALUES
  ('{{ $identity_id }}', '{{ .AppID }}', 'username', 'e2e_stored', '{"preferred_username": "e2e_stored"}', 'e2e_stored', 'e2e_stored', 'username');
