INSERT INTO _auth_resource (
  "id",
  "app_id",
  "created_at",
  "updated_at",
  "uri",
  "name",
  "metadata"
) VALUES (
  '{{ .AppID }}-R1',
  '{{ .AppID }}',
  NOW(),
  NOW(),
  'https://r/1',
  'R1',
  '{}'
);

INSERT INTO _auth_resource_scope (
  "id",
  "app_id",
  "created_at",
  "updated_at",
  "resource_id",
  "scope",
  "description",
  "metadata"
) VALUES (
  '{{ .AppID }}-R1:S1',
  '{{ .AppID }}',
  NOW(),
  NOW(),
  '{{ .AppID }}-R1',
  'R1:S1',
  'R1 Scope 1',
  '{}'
);

INSERT INTO _auth_client_resource_scope (
  "id",
  "app_id",
  "created_at",
  "updated_at",
  "client_id",
  "resource_id",
  "scope_id"
) VALUES (
  '{{ uuidv4 }}',
  '{{ .AppID }}',
  NOW(),
  NOW(),
  'e2em2mclient',
  '{{ .AppID }}-R1',
  '{{ .AppID }}-R1:S1'
);
