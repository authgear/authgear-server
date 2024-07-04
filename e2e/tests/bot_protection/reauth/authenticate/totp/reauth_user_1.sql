INSERT INTO _auth_user ("id", "app_id", "created_at", "updated_at", "last_login_at", "login_at", "is_disabled", "disable_reason", "standard_attributes", "custom_attributes", "is_deactivated", "delete_at", "is_anonymized", "anonymize_at", "anonymized_at", "last_indexed_at", "require_reindex_after") VALUES
('f30f554a-4c3a-45b9-9895-d458a9d206dd', '{{ .AppID }}', '2024-07-04 18:27:05.456749', '2024-07-04 18:27:05.465868', NULL, NULL, 'f', NULL, '{"name": "TOTP reauth with bot protection", "email": "bpreauth_authn_totp@example.com"}', '{}', 'f', NULL, 'f', NULL, NULL, '2024-07-04 18:27:05.492059', '2024-07-04 18:27:05.470143');

INSERT INTO _auth_authenticator ("id", "app_id", "type", "user_id", "created_at", "updated_at", "is_default", "kind") VALUES
('81c6f842-a312-40c4-8245-bc0a84ce3db0', '{{ .AppID }}', 'password', 'f30f554a-4c3a-45b9-9895-d458a9d206dd', '2024-07-04 18:27:05.46671', '2024-07-04 18:27:05.46671', 'f', 'primary'),
('c36ce39e-8e5b-4686-b874-d07a91bb0e5a', '{{ .AppID }}', 'totp', 'f30f554a-4c3a-45b9-9895-d458a9d206dd', '2024-07-04 18:27:05.469327', '2024-07-04 18:27:05.469327', 'f', 'secondary');

INSERT INTO _auth_authenticator_password ("id", "app_id", "password_hash", "expire_after") VALUES
('81c6f842-a312-40c4-8245-bc0a84ce3db0', '{{ .AppID }}', '$2y$10$/wLavnCmYGP/zzpw/mR1iOK5y5hGyrEFJtmaIbvFf9VA6l2O4NMKO', NULL);

INSERT INTO _auth_authenticator_totp ("id", "app_id", "secret", "display_name") VALUES
('c36ce39e-8e5b-4686-b874-d07a91bb0e5a', '{{ .AppID }}', '3I526Y3Y7GSXO34RTFEEFXCJM6Y4VZXR', 'Imported');

INSERT INTO _auth_identity ("id", "app_id", "type", "user_id", "created_at", "updated_at") VALUES
('b07cce43-4d60-4723-ab1c-9e0fac13e31c', '{{ .AppID }}', 'login_id', 'f30f554a-4c3a-45b9-9895-d458a9d206dd', '2024-07-04 18:27:05.459533', '2024-07-04 18:27:05.459533');

INSERT INTO _auth_identity_login_id ("id", "app_id", "login_id_key", "login_id", "claims", "original_login_id", "unique_key", "login_id_type") VALUES
('b07cce43-4d60-4723-ab1c-9e0fac13e31c', '{{ .AppID }}', 'email', 'bpreauth_authn_totp@example.com', '{"email": "bpreauth_authn_totp@example.com"}', 'bpreauth_authn_totp@example.com', 'bpreauth_authn_totp@example.com', 'email');

INSERT INTO _auth_password_history ("id", "app_id", "created_at", "user_id", "password") VALUES
('ab9703e8-9654-49be-ba70-d366f042d332', '{{ .AppID }}', '2024-07-04 18:27:05.467817', 'f30f554a-4c3a-45b9-9895-d458a9d206dd', '$2y$10$/wLavnCmYGP/zzpw/mR1iOK5y5hGyrEFJtmaIbvFf9VA6l2O4NMKO');

