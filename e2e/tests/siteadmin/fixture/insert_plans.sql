-- Seed three known plans used by the plan e2e tests.
-- ON CONFLICT DO NOTHING makes this idempotent across parallel test runs.

INSERT INTO _portal_plan (id, name, feature_config, created_at, updated_at)
VALUES
    (gen_random_uuid()::text, 'free',       '{}', NOW(), NOW()),
    (gen_random_uuid()::text, 'startups',   '{}', NOW(), NOW()),
    (gen_random_uuid()::text, 'enterprise', '{}', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;
