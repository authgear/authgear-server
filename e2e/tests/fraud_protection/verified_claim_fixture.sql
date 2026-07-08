-- Insert a minimal user so the verified claim foreign key is satisfied.
WITH new_user AS (
  INSERT INTO _auth_user (id, app_id, created_at, updated_at, is_disabled)
  VALUES ('{{ uuidv4 }}', '{{ .AppID }}', NOW(), NOW(), FALSE)
  RETURNING id
)
INSERT INTO _auth_verified_claim (id, app_id, user_id, name, value, created_at, metadata)
SELECT '{{ uuidv4 }}', '{{ .AppID }}', id, 'phone_number', '+6591230001', NOW(), NULL
FROM new_user;
