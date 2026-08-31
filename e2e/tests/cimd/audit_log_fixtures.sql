-- A pre-existing, STALE CIMD row (last_fetched_at 2 hours ago) so the
-- first authorize in audit_log.test.yaml triggers a refetch rather than a
-- fresh creation -- same shape as stale_record_refetch_fixtures.sql, kept
-- as its own copy so this file's audit-log assertions don't depend on a
-- fixture another test file could change independently.
INSERT INTO _auth_oauth_client (
  id, app_id, client_id, source, created_at, updated_at, last_fetched_at,
  kind, application_type, client_name, redirect_uris, grant_types, response_types
) VALUES (
  '{{ .AppID }}-cimd-audit-stale',
  '{{ .AppID }}',
  'http://localhost:2727/audit-log-stale.json',
  'CIMD',
  NOW() - interval '2 hours', NOW() - interval '2 hours', NOW() - interval '2 hours',
  'THIRD_PARTY', 'web', 'Audit Stale Client',
  '["http://127.0.0.1:9000/cb"]', '["authorization_code","refresh_token"]', '["code"]'
);
