-- Resource for client-scope tests
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

-- Scopes for the resource
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
  '{{ .AppID }}-resource2_scope1',
  '{{ .AppID }}',
  NOW(),
  NOW(),
  '{{ .AppID }}-fixture-resource-02',
  'resource2_scope1',
  'Resource 2 Scope 1',
  '{}'
), (
  '{{ .AppID }}-resource2_scope2',
  '{{ .AppID }}',
  NOW(),
  NOW(),
  '{{ .AppID }}-fixture-resource-02',
  'resource2_scope2',
  'Resource 2 Scope 2',
  '{}'
);

-- Associate the resource with the client 'portal'
INSERT INTO _auth_client_resource (
  "id",
  "app_id",
  "created_at",
  "updated_at",
  "client_id",
  "resource_id"
) VALUES (
  '{{ uuidv4 }}',
  '{{ .AppID }}',
  NOW(),
  NOW(),
  'portal',
  '{{ .AppID }}-fixture-resource-02'
);

-- Associate both scopes with the client 'portal' for resource 2
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
  'portal',
  '{{ .AppID }}-fixture-resource-02',
  '{{ .AppID }}-resource2_scope1'
), (
  '{{ uuidv4 }}',
  '{{ .AppID }}',
  NOW(),
  NOW(),
  'portal',
  '{{ .AppID }}-fixture-resource-02',
  '{{ .AppID }}-resource2_scope2'
); 
