with a as (
  SELECT login_id, _auth_authenticator.id as authenticator_id
  FROM _auth_authenticator
  JOIN _auth_user ON _auth_authenticator.user_id = _auth_user.id
  JOIN _auth_identity ON _auth_authenticator.user_id = _auth_user.id
  JOIN _auth_identity_login_id ON _auth_identity_login_id.id = _auth_identity.id
)
update _auth_authenticator
set updated_at = '2000-01-01 10:10:10'
from a
WHERE a.login_id = 'e2e_recipe_1_expiry@authgear.com'
AND id = a.authenticator_id
AND app_id = '{{ .AppID }}';
