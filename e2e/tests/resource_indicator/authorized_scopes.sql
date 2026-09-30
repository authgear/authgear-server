-- Fixed IDs, so that the Admin API lists the two Resources in a known order.
INSERT INTO _auth_resource (id, app_id, created_at, updated_at, uri, access_policy)
VALUES
  ('{{ .AppID }}-a-orders', '{{ .AppID }}', NOW(), NOW(), 'https://api.example.com/orders', '{"allow_static_third_party_client_access": true}'),
  ('{{ .AppID }}-b-inventory', '{{ .AppID }}', NOW(), NOW(), 'https://api.example.com/inventory', '{"allow_static_third_party_client_access": true}');

INSERT INTO _auth_resource_scope (id, app_id, created_at, updated_at, resource_id, scope, description, access_policy)
VALUES
  ('{{ uuidv4 }}', '{{ .AppID }}', NOW(), NOW(), '{{ .AppID }}-a-orders', 'read:orders', 'Read your orders', '{"allow_static_third_party_client_access": true}'),
  ('{{ uuidv4 }}', '{{ .AppID }}', NOW(), NOW(), '{{ .AppID }}-b-inventory', 'read:orders', 'Read your stock', '{"allow_static_third_party_client_access": true}');
