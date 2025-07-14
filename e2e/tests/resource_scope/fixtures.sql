INSERT INTO _auth_resource (
  "id",
  "app_id",
  "created_at",
  "updated_at",
  "uri",
  "name",
  "metadata"
) VALUES (
  '{{ .AppID }}-fixture-resource-01',
  '{{ .AppID }}',
  NOW(),
  NOW(),
  'https://fixtureresource/1',
  'Fixture Resource 1',
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
  '{{ .AppID }}-fixture-scope-01',
  '{{ .AppID }}',
  NOW(),
  NOW(),
  '{{ .AppID }}-fixture-resource-01',
  'fixturescope1',
  'Fixture Scope 1',
  '{}'
);
