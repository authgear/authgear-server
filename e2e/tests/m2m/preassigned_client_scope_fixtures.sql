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

-- Associate the resource with the client 'e2econfidential'
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
  'e2econfidential',
  '{{ .AppID }}-fixture-resource-02'
);


-- Associate both scopes with the client 'e2econfidential' for resource 2
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
  'e2econfidential',
  '{{ .AppID }}-fixture-resource-02',
  '{{ .AppID }}-resource2_scope1'
), (
  '{{ uuidv4 }}',
  '{{ .AppID }}',
  NOW(),
  NOW(),
  'e2econfidential',
  '{{ .AppID }}-fixture-resource-02',
  '{{ .AppID }}-resource2_scope2'
);

-- Associate the resource with the client 'e2em2mclient'
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
  'e2em2mclient',
  '{{ .AppID }}-fixture-resource-02'
);

-- Associate both scopes with the client 'e2em2mclient' for resource 2
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
  '{{ .AppID }}-fixture-resource-02',
  '{{ .AppID }}-resource2_scope1'
), (
  '{{ uuidv4 }}',
  '{{ .AppID }}',
  NOW(),
  NOW(),
  'e2em2mclient',
  '{{ .AppID }}-fixture-resource-02',
  '{{ .AppID }}-resource2_scope2'
);
