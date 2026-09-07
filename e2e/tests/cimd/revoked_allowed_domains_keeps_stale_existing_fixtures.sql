-- A pre-existing, STALE CIMD row (last_fetched_at 2 hours ago, past the
-- 1 hour refetch interval) on a host that is NOT in the project's
-- allowed_domains (localhost is not *.example.com). Being stale, this row
-- WOULD normally be refetched -- so this fixture, unlike
-- revoked_allowed_domains_keeps_existing_fixtures.sql's fresh row, actually
-- exercises the domain-trust check on the refetch path.
INSERT INTO _auth_oauth_client (
  id, app_id, client_id, source, created_at, updated_at, last_fetched_at,
  kind, application_type, client_name, redirect_uris, grant_types, response_types
) VALUES (
  '{{ .AppID }}-cimd-seeded-stale',
  '{{ .AppID }}',
  'http://localhost:2727/revoked-domains-stale.json',
  'CIMD',
  NOW() - interval '2 hours', NOW() - interval '2 hours', NOW() - interval '2 hours',
  'THIRD_PARTY', 'web', 'Seeded Stale CIMD Client',
  '["http://127.0.0.1:9000/cb"]', '["authorization_code","refresh_token"]', '["code"]'
);
