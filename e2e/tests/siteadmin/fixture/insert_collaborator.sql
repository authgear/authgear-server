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
