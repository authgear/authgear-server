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

-- Fixture for addScopesToClientID test
INSERT INTO _auth_resource (
  "id",
  "app_id",
  "created_at",
  "updated_at",
  "uri",
  "name",
  "metadata"
) VALUES (
  '{{ .AppID }}-fixture-resource-02',
  '{{ .AppID }}',
  NOW(),
  NOW(),
  'https://fixtureresource/2',
  'Fixture Resource 2',
  '{}'
);

INSERT INTO _auth_client_resource (
  "id",
  "app_id",
  "created_at",
  "updated_at",
  "client_id",
  "resource_id"
) VALUES (
  '{{ .AppID }}-fixture-client-resource-01',
  '{{ .AppID }}',
  NOW(),
  NOW(),
  'portal',
  '{{ .AppID }}-fixture-resource-02'
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
  '{{ .AppID }}-fixture-scope-02',
  '{{ .AppID }}',
  NOW(),
  NOW(),
  '{{ .AppID }}-fixture-resource-02',
  'fixturescope2',
  'Fixture Scope 2',
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
  '{{ .AppID }}-fixture-scope-03',
  '{{ .AppID }}',
  NOW(),
  NOW(),
  '{{ .AppID }}-fixture-resource-02',
  'fixturescope3',
  'Fixture Scope 3',
  '{}'
);
