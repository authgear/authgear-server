-- A pre-existing DCR row, as if it had been registered while DCR was
-- enabled. DCR clients have no refetch concept (last_fetched_at is always
-- NULL for source = 'DCR'), so there is nothing to "freeze" here -- the
-- only question is whether it still resolves once DCR is disabled.
INSERT INTO _auth_oauth_client (
  id, app_id, client_id, source, created_at, updated_at, last_fetched_at,
  kind, application_type, client_name, redirect_uris, grant_types, response_types
) VALUES (
  '{{ .AppID }}-dcr-seeded',
  '{{ .AppID }}',
  'dcrc_disabled_keeps_existing_test',
  'DCR',
  NOW(), NOW(), NULL,
  'THIRD_PARTY', 'web', 'Seeded DCR Client',
  '["http://127.0.0.1:9000/cb"]', '["authorization_code","refresh_token"]', '["code"]'
);
