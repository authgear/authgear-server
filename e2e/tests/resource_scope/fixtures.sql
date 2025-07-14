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
