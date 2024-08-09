SELECT *
FROM _auth_user 
WHERE app_id = '{{ .AppID }}'
AND standard_attributes ->> 'preferred_username' = 'lowerUPPER';
