{{ $user_id := (uuidv4) }}
{{ $authenticator_password_id := (uuidv4) }}
{{ $identity_id_1 := (uuidv4) }}
{{ $identity_id_2 := (uuidv4) }}
{{ $password_history_id := (uuidv4) }}

INSERT INTO _auth_user ("id", "app_id", "created_at", "updated_at", "last_login_at", "login_at", "is_disabled", "disable_reason", "standard_attributes", "custom_attributes", "is_deactivated", "delete_at", "is_anonymized", "anonymize_at", "anonymized_at", "last_indexed_at", "require_reindex_after") VALUES
('{{ $user_id }}', '{{ .AppID }}', '2024-06-27 07:51:42.040683', '2024-06-27 07:51:42.056654', NULL, NULL, 'f', NULL, '{"email": "e2e_reauth_primary_password@example.com", "preferred_username": "e2e_reauth_primary_password"}', '{}', 'f', NULL, 'f', NULL, NULL, '2024-06-27 07:51:42.079532', '2024-06-27 07:51:42.059056');

INSERT INTO _auth_authenticator ("id", "app_id", "type", "user_id", "created_at", "updated_at", "is_default", "kind") VALUES
('{{ $authenticator_password_id }}', '{{ .AppID }}', 'password', '{{ $user_id }}', '2024-06-27 07:51:42.057486', '2024-06-27 07:51:42.057486', 't', 'primary');

INSERT INTO
  _auth_authenticator_password ("id", "app_id", "password_hash", "expire_after")
VALUES
  (
    '{{ $authenticator_password_id }}',
    '{{ .AppID }}',
    '$bcrypt-sha512$$2a$10$20bGr8PamvoeIhOmPAEcDuSnWTcBDTJ/eJFh/SwAeVeUBCBhbx2u2',
    NULL
  );

INSERT INTO _auth_identity ("id", "app_id", "type", "user_id", "created_at", "updated_at") VALUES
('{{ $identity_id_1 }}', '{{ .AppID }}', 'login_id', '{{ $user_id }}', '2024-06-27 07:51:42.051107', '2024-06-27 07:51:42.051107'),
('{{ $identity_id_2 }}', '{{ .AppID }}', 'login_id', '{{ $user_id }}', '2024-06-27 07:51:42.046545', '2024-06-27 07:51:42.046545');

INSERT INTO _auth_identity_login_id ("id", "app_id", "login_id_key", "login_id", "claims", "original_login_id", "unique_key", "login_id_type") VALUES
('{{ $identity_id_1 }}', '{{ .AppID }}', 'username', 'e2e_reauth_primary_password', '{"preferred_username": "e2e_reauth_primary_password"}', 'e2e_reauth_primary_password', 'e2e_reauth_primary_password', 'username'),
('{{ $identity_id_2 }}', '{{ .AppID }}', 'email', 'e2e_reauth_primary_password@example.com', '{"email": "e2e_reauth_primary_password@example.com"}', 'e2e_reauth_primary_password@example.com', 'e2e_reauth_primary_password@example.com', 'email');

INSERT INTO _auth_password_history ("id", "app_id", "created_at", "user_id", "password") VALUES
('{{ $password_history_id }}', '{{ .AppID }}', '2024-06-27 07:51:42.058464', '{{ $user_id }}', '$2y$10$/wLavnCmYGP/zzpw/mR1iOK5y5hGyrEFJtmaIbvFf9VA6l2O4NMKO');



