-- Shared Site Admin seed used by the app, collaborator, plan, and usage e2e tests.
-- It grants Site Admin access in e2e-portal, inserts known test apps, and seeds
-- a known owner user for e2e-siteadmin-app-alpha.

INSERT INTO _portal_app_collaborator (id, app_id, user_id, created_at, role)
VALUES (
    gen_random_uuid()::text,
    'e2e-portal',
    '00000000-0000-0000-0000-000000000001',
    NOW(),
    'owner'
)
ON CONFLICT (app_id, user_id) DO NOTHING;

INSERT INTO _portal_config_source (id, app_id, data, plan_name, created_at, updated_at)
VALUES
    (
        gen_random_uuid()::text,
        'e2e-siteadmin-app-alpha',
        '{}',
        'startups',
        NOW() - INTERVAL '2 months',
        NOW() - INTERVAL '2 months'
    ),
    (
        gen_random_uuid()::text,
        'e2e-siteadmin-app-beta',
        '{}',
        'free',
        NOW() - INTERVAL '1 month',
        NOW() - INTERVAL '1 month'
    ),
    (
        gen_random_uuid()::text,
        'e2e-siteadmin-app-gamma',
        '{}',
        'startups',
        NOW() - INTERVAL '1 month',
        NOW() - INTERVAL '1 month'
    )
ON CONFLICT (app_id) DO NOTHING;

-- Insert last-month MAU usage records for sorting/filtering tests.
INSERT INTO _portal_usage_record (id, app_id, name, period, start_time, end_time, count)
VALUES
    (
        gen_random_uuid()::text,
        'e2e-siteadmin-app-alpha',
        'active-user',
        'monthly',
        date_trunc('month', (NOW() AT TIME ZONE 'UTC') - INTERVAL '1 month'),
        date_trunc('month', NOW() AT TIME ZONE 'UTC'),
        100
    ),
    (
        gen_random_uuid()::text,
        'e2e-siteadmin-app-beta',
        'active-user',
        'monthly',
        date_trunc('month', (NOW() AT TIME ZONE 'UTC') - INTERVAL '1 month'),
        date_trunc('month', NOW() AT TIME ZONE 'UTC'),
        500
    ),
    (
        gen_random_uuid()::text,
        'e2e-siteadmin-app-gamma',
        'active-user',
        'monthly',
        date_trunc('month', (NOW() AT TIME ZONE 'UTC') - INTERVAL '1 month'),
        date_trunc('month', NOW() AT TIME ZONE 'UTC'),
        200
    )
ON CONFLICT (app_id, name, period, start_time) DO UPDATE SET count = EXCLUDED.count;

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
    '{"email": "owner@example.com"}',
    '{}',
    false, NULL, false, NULL, NULL,
    NOW(), NOW()
)
ON CONFLICT (id) DO NOTHING;

-- Add a login ID identity so Admin API getUsersByStandardAttribute(email, ...)
-- can resolve this user by email, not just by direct node lookup.
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

INSERT INTO _portal_app_collaborator (id, app_id, user_id, created_at, role)
VALUES (
    gen_random_uuid()::text,
    'e2e-siteadmin-app-alpha',
    '00000000-0000-0000-0000-000000000002',
    NOW(),
    'owner'
)
ON CONFLICT (app_id, user_id) DO NOTHING;
