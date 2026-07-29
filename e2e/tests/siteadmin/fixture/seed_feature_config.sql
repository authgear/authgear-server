-- Dedicated plan + app for feature config e2e tests, isolated from the
-- apps/plans/collaborators fixtures so this test can freely mutate the
-- app's feature config override without affecting other suites.
INSERT INTO _portal_plan (id, name, feature_config, created_at, updated_at)
VALUES (
    gen_random_uuid()::text,
    'e2e-feature-config-plan',
    '{"collaborator": {"maximum": 3}, "oauth": {"client": {"maximum": 10}}}',
    NOW(), NOW()
)
ON CONFLICT (name) DO NOTHING;

INSERT INTO _portal_config_source (id, app_id, data, plan_name, created_at, updated_at)
VALUES (
    gen_random_uuid()::text,
    'e2e-feature-config-app',
    '{}',
    'e2e-feature-config-plan',
    NOW(), NOW()
)
ON CONFLICT (app_id) DO NOTHING;
