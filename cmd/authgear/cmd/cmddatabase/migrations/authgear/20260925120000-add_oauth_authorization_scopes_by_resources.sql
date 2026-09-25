-- +migrate Up
-- The server no longer writes scopes. The column is kept for an older binary.
ALTER TABLE _auth_oauth_authorization ALTER COLUMN scopes SET DEFAULT '[]'::jsonb;

-- The default lets an older binary still running during rollout insert rows.
ALTER TABLE _auth_oauth_authorization ADD COLUMN scopes_by_resources jsonb NOT NULL DEFAULT '{}'::jsonb;

-- Project-level scopes go under "authgear". A resource scope is granted on every
-- resource that defines it, since the recorded name alone cannot tell which one
-- it was granted for. A name no resource defines any more is dropped.
UPDATE _auth_oauth_authorization a SET scopes_by_resources = COALESCE((
  SELECT jsonb_object_agg(g.key, g.scopes) FROM (
    SELECT 'authgear' AS key, jsonb_agg(t.scope ORDER BY t.ord) AS scopes
    FROM jsonb_array_elements_text(a.scopes) WITH ORDINALITY AS t(scope, ord)
    WHERE t.scope IN (
      'offline_access',
      'device_sso',
      'openid',
      'profile',
      'email',
      'address',
      'phone',
      'https://authgear.com/scopes/full-access',
      'https://authgear.com/scopes/full-userinfo',
      'https://authgear.com/scopes/pre-authenticated-url'
    )
    HAVING count(*) > 0
    UNION ALL
    SELECT rs.resource_id AS key, jsonb_agg(t.scope ORDER BY t.ord) AS scopes
    FROM jsonb_array_elements_text(a.scopes) WITH ORDINALITY AS t(scope, ord)
    JOIN _auth_resource_scope rs ON rs.app_id = a.app_id AND rs.scope = t.scope
    GROUP BY rs.resource_id
  ) g
), '{}'::jsonb);

-- +migrate Down
ALTER TABLE _auth_oauth_authorization DROP COLUMN scopes_by_resources;
ALTER TABLE _auth_oauth_authorization ALTER COLUMN scopes DROP DEFAULT;
