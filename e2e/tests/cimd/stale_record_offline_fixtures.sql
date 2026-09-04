-- A pre-existing CIMD row, STALE (last_fetched_at 2 hours ago, past the
-- 1 hour refetch interval), so the authorize request that follows will
-- attempt (and, per the test's document-server configuration, fail) a
-- refetch.
INSERT INTO _auth_oauth_client (
  id, app_id, client_id, source, created_at, updated_at, last_fetched_at,
  kind, application_type, client_name, redirect_uris, grant_types, response_types
) VALUES (
  '{{ .AppID }}-cimd-stale',
  '{{ .AppID }}',
  'http://localhost:2727/stale-record-offline.json',
  'CIMD',
  NOW() - interval '2 hours', NOW() - interval '2 hours', NOW() - interval '2 hours',
  'THIRD_PARTY', 'web', 'Stale CIMD Client',
  '["http://127.0.0.1:9000/cb"]', '["authorization_code","refresh_token"]', '["code"]'
);
