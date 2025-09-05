INSERT INTO _auth_recovery_code (
  "id",
  "user_id",
  "app_id",
  "code",
  "created_at",
  "updated_at",
  "consumed"
) VALUES (
  '{{ uuidv4 }}',
  (SELECT i.user_id
    FROM _auth_identity_login_id a
    JOIN _auth_identity i ON a.id = i.id
    WHERE a.login_id = 'e2e_recovery_code_user_only_recovery_code@example.com'
    AND i.app_id = '{{ .AppID }}'
    AND a.login_id_type = 'email'
    LIMIT 1
  ),
  '{{ .AppID }}',
  'RECOVERYCODE1',
  NOW(),
  NOW(),
  FALSE
);

INSERT INTO _auth_recovery_code (
 "id",
 "user_id",
 "app_id",
 "code",
 "created_at",
 "updated_at",
 "consumed"
) VALUES (
 '{{ uuidv4 }}',
 (SELECT i.user_id
   FROM _auth_identity_login_id a
   JOIN _auth_identity i ON a.id = i.id
   WHERE a.login_id = 'e2e_recovery_code_user_with_mfa_email@example.com'
   AND i.app_id = '{{ .AppID }}'
   AND a.login_id_type = 'email'
   LIMIT 1
 ),
 '{{ .AppID }}',
 'RECOVERYCODE3',
 NOW(),
 NOW(),
 FALSE
);

INSERT INTO _auth_recovery_code (
  "id",
  "user_id",
  "app_id",
  "code",
  "created_at",
  "updated_at",
  "consumed"
) VALUES (
  '{{ uuidv4 }}',
  (SELECT i.user_id
    FROM _auth_identity_login_id a
    JOIN _auth_identity i ON a.id = i.id
    WHERE a.login_id = 'e2e_recovery_code_user_with_sms@example.com'
    AND a.app_id = '{{ .AppID }}'
    AND a.login_id_type = 'email'
    LIMIT 1
  ),
  '{{ .AppID }}',
  'RECOVERYCODE2',
  NOW(),
  NOW(),
  FALSE
);
