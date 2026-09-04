-- A pre-existing, FRESH CIMD row on a host that is NOT in the project's
-- allowed_domains (localhost is not *.example.com). allowed_domains is
-- checked only on the fetch path (Part 1 §4.1); this row's freshness means
-- EnsureClientResolved never reaches that check for it.
INSERT INTO _auth_oauth_client (
  id, app_id, client_id, source, created_at, updated_at, last_fetched_at,
  kind, application_type, client_name, redirect_uris, grant_types, response_types
) VALUES (
  '{{ .AppID }}-cimd-seeded',
  '{{ .AppID }}',
  'http://localhost:2727/revoked-domains-seeded.json',
  'CIMD',
  NOW(), NOW(), NOW(),
  'THIRD_PARTY', 'web', 'Seeded CIMD Client',
  '["http://127.0.0.1:9000/cb"]', '["authorization_code","refresh_token"]', '["code"]'
);
