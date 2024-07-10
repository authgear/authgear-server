INSERT INTO _auth_user ("id", "app_id", "created_at", "updated_at", "last_login_at", "login_at", "is_disabled", "disable_reason", "standard_attributes", "custom_attributes", "is_deactivated", "delete_at", "is_anonymized", "anonymize_at", "anonymized_at", "last_indexed_at", "require_reindex_after") VALUES
('2a06e4bd-832f-4f14-a693-403470117f1f', '{{ .AppID }}', '2024-07-04 17:51:02.786609', '2024-07-04 17:51:02.799511', NULL, NULL, 'f', NULL, '{"email": "bpreauth_authn_oobotp@example.com"}', '{}', 'f', NULL, 'f', NULL, NULL, '2024-07-04 17:51:02.826179', '2024-07-04 17:51:02.803398');

INSERT INTO _auth_identity ("id", "app_id", "type", "user_id", "created_at", "updated_at") VALUES
('2f75553d-d073-4e12-8d33-5e773a5ca526', '{{ .AppID }}', 'login_id', '2a06e4bd-832f-4f14-a693-403470117f1f', '2024-07-04 17:51:02.791893', '2024-07-04 17:51:02.791893');

INSERT INTO _auth_authenticator ("id", "app_id", "type", "user_id", "created_at", "updated_at", "is_default", "kind") VALUES
('3b4f6c3e-3a42-4557-bf73-9bb81dbc2c62', '{{ .AppID }}', 'password', '2a06e4bd-832f-4f14-a693-403470117f1f', '2024-07-04 17:51:02.800397', '2024-07-04 17:51:02.800397', 'f', 'primary'),
('fdd6dfc3-73fa-4370-84a9-5187a125163b', '{{ .AppID }}', 'oob_otp_email', '2a06e4bd-832f-4f14-a693-403470117f1f', '2024-07-04 17:51:02.802635', '2024-07-04 17:51:02.802635', 'f', 'secondary');

INSERT INTO _auth_authenticator_oob ("id", "app_id", "phone", "email") VALUES
('fdd6dfc3-73fa-4370-84a9-5187a125163b', '{{ .AppID }}', '', 'bpreauth_authn_oobotp@example.com');

INSERT INTO _auth_authenticator_password ("id", "app_id", "password_hash", "expire_after") VALUES
('3b4f6c3e-3a42-4557-bf73-9bb81dbc2c62', '{{ .AppID }}', '$2y$10$ze9h.zFQCg9ew9X4orQE2OkuLJZvTDs/YwOuptGkzkiFvwsES16hK', NULL);

INSERT INTO _auth_identity_login_id ("id", "app_id", "login_id_key", "login_id", "claims", "original_login_id", "unique_key", "login_id_type") VALUES
('2f75553d-d073-4e12-8d33-5e773a5ca526', '{{ .AppID }}', 'email', 'bpreauth_authn_oobotp@example.com', '{"email": "bpreauth_authn_oobotp@example.com"}', 'bpreauth_authn_oobotp@example.com', 'bpreauth_authn_oobotp@example.com', 'email');

INSERT INTO _auth_password_history ("id", "app_id", "created_at", "user_id", "password") VALUES
('7037dd05-0938-4351-acc9-931ce1e2e17a', '{{ .AppID }}', '2024-07-04 17:51:02.801309', '2a06e4bd-832f-4f14-a693-403470117f1f', '$2y$10$ze9h.zFQCg9ew9X4orQE2OkuLJZvTDs/YwOuptGkzkiFvwsES16hK');