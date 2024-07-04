INSERT INTO _auth_user ("id", "app_id", "created_at", "updated_at", "last_login_at", "login_at", "is_disabled", "disable_reason", "standard_attributes", "custom_attributes", "is_deactivated", "delete_at", "is_anonymized", "anonymize_at", "anonymized_at", "last_indexed_at", "require_reindex_after") VALUES
('ecaad8f0-74aa-4d6f-8d7e-f4edcb0c43c8', '{{ .AppID }}', '2024-06-27 07:51:42.040683', '2024-06-27 07:51:42.056654', NULL, NULL, 'f', NULL, '{"email": "e2e_reauth_primary_password@example.com", "preferred_username": "e2e_reauth_primary_password"}', '{}', 'f', NULL, 'f', NULL, NULL, '2024-06-27 07:51:42.079532', '2024-06-27 07:51:42.059056') ON CONFLICT DO NOTHING;

INSERT INTO _auth_authenticator ("id", "app_id", "type", "user_id", "created_at", "updated_at", "is_default", "kind") VALUES
('70f573c9-90e6-4d5e-8b94-5cb6b4331fda', '{{ .AppID }}', 'password', 'ecaad8f0-74aa-4d6f-8d7e-f4edcb0c43c8', '2024-06-27 07:51:42.057486', '2024-06-27 07:51:42.057486', 't', 'primary') ON CONFLICT DO NOTHING;

INSERT INTO
  _auth_authenticator_password ("id", "app_id", "password_hash", "expire_after")
VALUES
  (
    '70f573c9-90e6-4d5e-8b94-5cb6b4331fda',
    '{{ .AppID }}',
    '$bcrypt-sha512$$2a$10$20bGr8PamvoeIhOmPAEcDuSnWTcBDTJ/eJFh/SwAeVeUBCBhbx2u2',
    NULL
  ) ON CONFLICT DO NOTHING;

INSERT INTO _auth_identity ("id", "app_id", "type", "user_id", "created_at", "updated_at") VALUES
('258c5774-47f9-4cad-baa2-dbaeac59e6f8', '{{ .AppID }}', 'login_id', 'ecaad8f0-74aa-4d6f-8d7e-f4edcb0c43c8', '2024-06-27 07:51:42.051107', '2024-06-27 07:51:42.051107'),
('4a0665ff-b910-49b4-ab17-cd1c7c18de3e', '{{ .AppID }}', 'login_id', 'ecaad8f0-74aa-4d6f-8d7e-f4edcb0c43c8', '2024-06-27 07:51:42.046545', '2024-06-27 07:51:42.046545') ON CONFLICT DO NOTHING;

INSERT INTO _auth_identity_login_id ("id", "app_id", "login_id_key", "login_id", "claims", "original_login_id", "unique_key", "login_id_type") VALUES
('258c5774-47f9-4cad-baa2-dbaeac59e6f8', '{{ .AppID }}', 'username', 'e2e_reauth_primary_password', '{"preferred_username": "e2e_reauth_primary_password"}', 'e2e_reauth_primary_password', 'e2e_reauth_primary_password', 'username'),
('4a0665ff-b910-49b4-ab17-cd1c7c18de3e', '{{ .AppID }}', 'email', 'e2e_reauth_primary_password@example.com', '{"email": "e2e_reauth_primary_password@example.com"}', 'e2e_reauth_primary_password@example.com', 'e2e_reauth_primary_password@example.com', 'email') ON CONFLICT DO NOTHING;

INSERT INTO _auth_password_history ("id", "app_id", "created_at", "user_id", "password") VALUES
('00699609-d17f-471a-a0ed-07251c0fbb4f', '{{ .AppID }}', '2024-06-27 07:51:42.058464', 'ecaad8f0-74aa-4d6f-8d7e-f4edcb0c43c8', '$2y$10$/wLavnCmYGP/zzpw/mR1iOK5y5hGyrEFJtmaIbvFf9VA6l2O4NMKO') ON CONFLICT DO NOTHING;



