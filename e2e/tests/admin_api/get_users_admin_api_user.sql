{{ $email_user_id := (uuidv4) }}
{{ $email_identity_id := (uuidv4) }}
{{ $phone_number_user_id := (uuidv4) }}
{{ $phone_number_identity_id := (uuidv4) }}
{{ $username_user_id := (uuidv4) }}
{{ $username_identity_id := (uuidv4) }}

INSERT INTO _auth_user
  ("id", "app_id", "created_at", "updated_at", "last_login_at", "login_at", "is_disabled", "disable_reason", "standard_attributes", "custom_attributes", "is_deactivated", "delete_at", "is_anonymized", "anonymize_at", "anonymized_at", "last_indexed_at", "require_reindex_after")
VALUES
  ('{{ $email_user_id }}', '{{ .AppID }}', '2024-08-27 07:51:42.040683', '2024-08-27 07:51:42.056654', NULL, NULL, 'f', NULL, '{"email": "e2e_admin_api_get_users@example.com"}', '{}', 'f', NULL, 'f', NULL, NULL, '2024-08-27 07:51:42.079532', '2024-08-27 07:51:42.059056'),
  ('{{ $phone_number_user_id }}', '{{ .AppID }}', '2024-08-27 07:51:42.040683', '2024-08-27 07:51:42.056654', NULL, NULL, 'f', NULL, '{"phone_number": "+85261236544"}', '{}', 'f', NULL, 'f', NULL, NULL, '2024-08-27 07:51:42.079532', '2024-08-27 07:51:42.059056'),
  ('{{ $username_user_id }}', '{{ .AppID }}', '2024-08-27 07:51:42.040683', '2024-08-27 07:51:42.056654', NULL, NULL, 'f', NULL, '{"preferred_username": "e2e_admin_api_get_users_username"}', '{}', 'f', NULL, 'f', NULL, NULL, '2024-08-27 07:51:42.079532', '2024-08-27 07:51:42.059056');

INSERT INTO _auth_identity
  ("id", "app_id", "type", "user_id", "created_at", "updated_at")
VALUES
  ('{{ $email_identity_id }}', '{{ .AppID }}', 'login_id', '{{ $email_user_id }}', '2024-08-27 07:51:42.051107', '2024-08-27 07:51:42.051107'),
  ('{{ $phone_number_identity_id }}', '{{ .AppID }}', 'login_id', '{{ $phone_number_user_id }}', '2024-08-27 07:51:42.051107', '2024-08-27 07:51:42.051107'),
  ('{{ $username_identity_id }}', '{{ .AppID }}', 'login_id', '{{ $username_user_id }}', '2024-08-27 07:51:42.051107', '2024-08-27 07:51:42.051107');


INSERT INTO _auth_identity_login_id
  ("id", "app_id", "login_id_key", "login_id", "claims", "original_login_id", "unique_key", "login_id_type")
VALUES
  ('{{ $email_identity_id }}', '{{ .AppID }}', 'email', 'e2e_admin_api_get_users@example.com', '{"email": "e2e_admin_api_get_users@example.com"}', 'e2e_admin_api_get_users@example.com', 'e2e_admin_api_get_users@example.com', 'email'),
  ('{{ $phone_number_identity_id }}', '{{ .AppID }}', 'phone', '+85261236544', '{"phone_number": "+85261236544"}', '+85261236544', '+85261236544', 'phone'),
  ('{{ $username_identity_id }}', '{{ .AppID }}', 'username', 'e2e_admin_api_get_users_username', '{"preferred_username": "e2e_admin_api_get_users_username"}', 'e2e_admin_api_get_users_username', 'e2e_admin_api_get_users_username', 'username');

