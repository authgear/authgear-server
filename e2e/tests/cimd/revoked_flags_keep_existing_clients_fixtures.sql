-- A pre-existing, FRESH CIMD row for an http:// client_id, as if it had
-- been resolved earlier while insecure_http_allowed was set. last_fetched_at
-- = NOW() means the freshness check (Part 3) short-circuits before ever
-- consulting insecure_http_allowed again -- proving that revoking the flag
-- does not affect a client that already has a persisted record.
INSERT INTO _auth_oauth_client (
  id, app_id, client_id, source, created_at, updated_at, last_fetched_at,
  kind, application_type, client_name, redirect_uris, grant_types, response_types
) VALUES (
  '{{ .AppID }}-cimd-seeded',
  '{{ .AppID }}',
  'http://localhost:2727/revoked-flags-seeded.json',
  'CIMD',
  NOW(), NOW(), NOW(),
  'THIRD_PARTY', 'web', 'Seeded CIMD Client',
  '["http://127.0.0.1:9000/cb"]', '["authorization_code","refresh_token"]', '["code"]'
);
