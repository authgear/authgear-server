-- Seed used exclusively by collaborators.test.yaml.
-- Uses dedicated app IDs (e2e-collab-*) so this test can freely mutate
-- collaborator rows without interfering with apps.test.yaml or other tests
-- that share e2e-siteadmin-app-*.

-- Grant Site Admin access in e2e-portal for the test actor.
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

-- Users (shared with other siteadmin seeds; ON CONFLICT keeps them idempotent).
INSERT INTO _auth_user (
    id, app_id, created_at, updated_at, last_login_at, login_at,
    is_disabled, disable_reason, standard_attributes, custom_attributes,
    is_deactivated, delete_at, is_anonymized, anonymize_at, anonymized_at,
    last_indexed_at, require_reindex_after
)
VALUES
    (
        '00000000-0000-0000-0000-000000000002',
        'e2e-portal',
        NOW(), NOW(), NULL, NULL,
        false, NULL,
        '{"email": "owner@example.com"}',
        '{}',
        false, NULL, false, NULL, NULL,
        NOW(), NOW()
    ),
    (
        '00000000-0000-0000-0000-000000000004',
        'e2e-portal',
        NOW(), NOW(), NULL, NULL,
        false, NULL,
        '{"email": "editor@example.com"}',
        '{}',
        false, NULL, false, NULL, NULL,
        NOW(), NOW()
    ),
    (
        '00000000-0000-0000-0000-000000000005',
        'e2e-portal',
        NOW(), NOW(), NULL, NULL,
        false, NULL,
        '{"email": "gamma-owner@example.com"}',
        '{}',
        false, NULL, false, NULL, NULL,
        NOW(), NOW()
    )
ON CONFLICT (id) DO NOTHING;

-- Login ID identity for user 002 so FindUserIDsByEmail("owner@example.com") works.
INSERT INTO _auth_identity (
    id, app_id, type, user_id, created_at, updated_at
)
VALUES (
    '00000000-0000-0000-0000-000000000003',
    'e2e-portal',
    'login_id',
    '00000000-0000-0000-0000-000000000002',
    NOW(),
    NOW()
)
ON CONFLICT DO NOTHING;

INSERT INTO _auth_identity_login_id (
    id, app_id, login_id_key, login_id, claims, original_login_id, unique_key, login_id_type
)
VALUES (
    '00000000-0000-0000-0000-000000000003',
    'e2e-portal',
    'email',
    'owner@example.com',
    '{"email": "owner@example.com"}',
    'owner@example.com',
    'owner@example.com',
    'email'
)
ON CONFLICT DO NOTHING;

-- Dedicated apps for collaborators.test.yaml — isolated from e2e-siteadmin-app-*.
INSERT INTO _portal_config_source (id, app_id, data, plan_name, created_at, updated_at)
VALUES
    (
        gen_random_uuid()::text,
        'e2e-collab-alpha',
        '{}',
        'free',
        NOW(),
        NOW()
    ),
    (
        gen_random_uuid()::text,
        'e2e-collab-beta',
        '{}',
        'free',
        NOW(),
        NOW()
    ),
    (
        gen_random_uuid()::text,
        'e2e-collab-gamma',
        '{}',
        'free',
        NOW(),
        NOW()
    )
ON CONFLICT (app_id) DO NOTHING;

-- Reset collaborators for all e2e-collab-* apps so previous runs don't leave
-- stale rows. This is safe because these app IDs are owned exclusively by this
-- seed file and collaborators.test.yaml.
DELETE FROM _portal_app_collaborator
WHERE app_id IN ('e2e-collab-alpha', 'e2e-collab-beta', 'e2e-collab-gamma');

-- e2e-collab-alpha: owner = user 005 (gamma-owner@example.com).
-- Deliberately NOT user 002 (owner@example.com) to avoid colliding with the
-- apps.test.yaml "filter by owner_email" step that expects exactly one result
-- for owner@example.com (e2e-siteadmin-app-alpha).
INSERT INTO _portal_app_collaborator (id, app_id, user_id, created_at, updated_at, role)
VALUES (
    gen_random_uuid()::text,
    'e2e-collab-alpha',
    '00000000-0000-0000-0000-000000000005',
    NOW(),
    NOW(),
    'owner'
);

-- e2e-collab-beta starts with no collaborators (test adds them dynamically).

-- e2e-collab-gamma: owner = user 005, editor = user 004.
INSERT INTO _portal_app_collaborator (id, app_id, user_id, created_at, updated_at, role)
VALUES
    (
        gen_random_uuid()::text,
        'e2e-collab-gamma',
        '00000000-0000-0000-0000-000000000005',
        NOW(), NOW(),
        'owner'
    ),
    (
        gen_random_uuid()::text,
        'e2e-collab-gamma',
        '00000000-0000-0000-0000-000000000004',
        NOW(), NOW(),
        'editor'
    );
