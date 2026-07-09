-- Run seed_siteadmin_actor.sql first to create user-001 and its e2e-portal membership.

-- User-002: valid auth user in e2e-portal but NOT a siteadmin collaborator.
-- Used by the unauthorized_user test to obtain a valid JWT that is rejected
-- by the authz middleware (403).
INSERT INTO _auth_user (
    id, app_id, created_at, updated_at, last_login_at, login_at,
    is_disabled, disable_reason, standard_attributes, custom_attributes,
    is_deactivated, delete_at, is_anonymized, anonymize_at, anonymized_at,
    last_indexed_at, require_reindex_after
)
VALUES (
    '00000000-0000-0000-0000-000000000002',
    'e2e-portal',
    NOW(), NOW(), NULL, NULL,
    false, NULL,
    '{"email": "nonmember@example.com"}',
    '{}',
    false, NULL, false, NULL, NULL,
    NOW(), NOW()
)
ON CONFLICT (id) DO NOTHING;
