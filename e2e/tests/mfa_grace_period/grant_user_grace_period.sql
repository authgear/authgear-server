with a as (
  SELECT _auth_user.id as user_id, _auth_identity_login_id.login_id as login_id
  FROM _auth_user
  JOIN _auth_identity ON _auth_identity.user_id = _auth_user.id
  JOIN _auth_identity_login_id ON _auth_identity_login_id.id = _auth_identity.id
)
UPDATE _auth_user
SET mfa_grace_period_end_at = '3000-01-01 10:10:10'
FROM a
WHERE a.login_id = 'e2e_mfa_grace_period@example.com'
AND id = a.user_id
AND app_id = '{{ .AppID }}';
