SELECT *
FROM _auth_user 
WHERE app_id = '{{ .AppID }}'
AND standard_attributes ->> 'email' = 'lowerUPPER@ca.se';
