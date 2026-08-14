-- Minimal seed: creates the siteadmin test actor (user-001) in the e2e-portal
-- auth database so they can obtain a valid Bearer JWT via the OAuth token
-- endpoint, and grants them siteadmin access in e2e-portal.

INSERT INTO _auth_user (
    id, app_id, created_at, updated_at, last_login_at, login_at,
    is_disabled, disable_reason, standard_attributes, custom_attributes,
    is_deactivated, delete_at, is_anonymized, anonymize_at, anonymized_at,
    last_indexed_at, require_reindex_after
)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'e2e-portal',
    NOW(), NOW(), NULL, NULL,
    false, NULL,
    '{"email": "siteadmin@example.com"}',
    '{}',
    false, NULL, false, NULL, NULL,
    NOW(), NOW()
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO _portal_app_collaborator (id, app_id, user_id, created_at, updated_at, role)
VALUES (
    gen_random_uuid()::text,
    'e2e-portal',
    '00000000-0000-0000-0000-000000000001',
    NOW(),
    NOW(),
    'owner'
)
ON CONFLICT (app_id, user_id) DO NOTHING;

-- Register localhost as a domain for e2e-portal so SITEADMIN_AUTHGEAR_ENDPOINT=http://localhost:4000
-- routes to e2e-portal without needing *.authgeare2e.localhost DNS resolution.
INSERT INTO _portal_domain (id, app_id, created_at, domain, apex_domain, verification_nonce, is_custom)
VALUES (
    '00000000-0000-0000-0000-000000000010',
    'e2e-portal',
    NOW(),
    'localhost',
    'localhost',
    'e2e-localhost-nonce',
    false
)
ON CONFLICT DO NOTHING;
