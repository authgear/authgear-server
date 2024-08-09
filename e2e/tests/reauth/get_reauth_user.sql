SELECT id
FROM _auth_user 
WHERE app_id = '{{ .AppID }}'
AND standard_attributes ->> 'preferred_username' = 'e2e_reauth_primary_password';
