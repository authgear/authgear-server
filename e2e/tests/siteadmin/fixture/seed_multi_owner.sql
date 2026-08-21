-- Regression fixture for the "an app can have more than one owner-role
-- collaborator" bug class (dev-3757): the schema only enforces
-- UNIQUE(app_id, user_id) on _portal_app_collaborator, nothing prevents two
-- rows for the same app both having role='owner' if that role is ever set
-- via a direct DB edit rather than through PromoteCollaborator. This seed
-- reproduces that state directly so the test can verify both call sites
-- that used to assume "at most one owner" handle it correctly:
--   1. ListAppsWithStats (app search/list) must not duplicate the app.
--   2. PromoteCollaborator must demote EVERY existing owner, not just one.
-- Uses a dedicated app ID so it can't interfere with apps.test.yaml or
-- collaborators.test.yaml.
-- Run seed_siteadmin_actor.sql first to create user-001 and its e2e-portal membership.

INSERT INTO _auth_user (
    id, app_id, created_at, updated_at, last_login_at, login_at,
    is_disabled, disable_reason, standard_attributes, custom_attributes,
    is_deactivated, delete_at, is_anonymized, anonymize_at, anonymized_at,
    last_indexed_at, require_reindex_after
)
VALUES
    (
        '00000000-0000-0000-0000-000000000101',
        'e2e-portal',
        NOW(), NOW(), NULL, NULL,
        false, NULL,
        '{"email": "multi-owner-1@example.com"}',
        '{}',
        false, NULL, false, NULL, NULL,
        NOW(), NOW()
    ),
    (
        '00000000-0000-0000-0000-000000000102',
        'e2e-portal',
        NOW(), NOW(), NULL, NULL,
        false, NULL,
        '{"email": "multi-owner-2@example.com"}',
        '{}',
        false, NULL, false, NULL, NULL,
        NOW(), NOW()
    ),
    (
        '00000000-0000-0000-0000-000000000103',
        'e2e-portal',
        NOW(), NOW(), NULL, NULL,
        false, NULL,
        '{"email": "multi-owner-editor@example.com"}',
        '{}',
        false, NULL, false, NULL, NULL,
        NOW(), NOW()
    )
ON CONFLICT (id) DO NOTHING;

INSERT INTO _portal_config_source (id, app_id, data, plan_name, created_at, updated_at)
VALUES (
    gen_random_uuid()::text,
    'e2e-multiowner-app',
    '{}',
    'free',
    NOW(),
    NOW()
)
ON CONFLICT (app_id) DO NOTHING;

-- Reset collaborators so previous runs don't leave stale rows -- this app ID
-- is owned exclusively by this seed file and multi_owner.test.yaml.
DELETE FROM _portal_app_collaborator WHERE app_id = 'e2e-multiowner-app';

-- Two owner-role rows for the SAME app -- the invalid-but-reachable state.
-- owner-1 is older so PromoteCollaborator's deterministic tie-break should
-- treat it as "primary".
INSERT INTO _portal_app_collaborator (id, app_id, user_id, created_at, updated_at, role)
VALUES
    (
        gen_random_uuid()::text,
        'e2e-multiowner-app',
        '00000000-0000-0000-0000-000000000101',
        NOW() - INTERVAL '1 hour',
        NOW() - INTERVAL '1 hour',
        'owner'
    ),
    (
        gen_random_uuid()::text,
        'e2e-multiowner-app',
        '00000000-0000-0000-0000-000000000102',
        NOW(),
        NOW(),
        'owner'
    ),
    (
        gen_random_uuid()::text,
        'e2e-multiowner-app',
        '00000000-0000-0000-0000-000000000103',
        NOW(),
        NOW(),
        'editor'
    );
