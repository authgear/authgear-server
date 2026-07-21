-- Seed site admin audit log entries for the audit log API e2e test.
-- Uses a dedicated affected app (e2e-audit-log-app) that no other test touches.
-- All entries are stored under app_id = 'e2e-portal' (the portal app),
-- matching how SiteAdminAuditService writes logs.
--
-- IDs are fixed hex sequence numbers; timestamps are spaced 1 hour apart
-- so ordering assertions are deterministic regardless of wall-clock time.

INSERT INTO _audit_log (id, app_id, created_at, user_id, activity_type, ip_address, user_agent, client_id, data)
VALUES
    (
        'e2e000000000001',
        'e2e-portal',
        '2026-01-01 10:00:00',
        '',
        'site_admin.app.plan.updated',
        '127.0.0.1',
        'Mozilla/5.0',
        '',
        '{
            "id": "e2e000000000001",
            "seq": 1,
            "type": "site_admin.app.plan.updated",
            "context": {
                "app_id": "e2e-portal",
                "user_id": null,
                "timestamp": 1751367600,
                "triggered_by": "site_admin",
                "ip_address": "127.0.0.1",
                "user_agent": "Mozilla/5.0",
                "audit_context": {
                    "usage": "internal",
                    "actor_user_id": "00000000-0000-0000-0000-000000000001",
                    "http_referer": "",
                    "http_url": "http://localhost:4003/api/v1/apps/e2e-audit-log-app/plan"
                },
                "preferred_languages": []
            },
            "payload": {
                "app_id": "e2e-audit-log-app",
                "old_plan": "free",
                "new_plan": "startups"
            }
        }'::jsonb
    ),
    (
        'e2e000000000002',
        'e2e-portal',
        '2026-01-01 11:00:00',
        '',
        'site_admin.app.collaborator.added',
        '127.0.0.1',
        'Mozilla/5.0',
        '',
        '{
            "id": "e2e000000000002",
            "seq": 2,
            "type": "site_admin.app.collaborator.added",
            "context": {
                "app_id": "e2e-portal",
                "user_id": null,
                "timestamp": 1751371200,
                "triggered_by": "site_admin",
                "ip_address": "127.0.0.1",
                "user_agent": "Mozilla/5.0",
                "audit_context": {
                    "usage": "internal",
                    "actor_user_id": "00000000-0000-0000-0000-000000000001",
                    "http_referer": "",
                    "http_url": "http://localhost:4003/api/v1/apps/e2e-audit-log-app/collaborators"
                },
                "preferred_languages": []
            },
            "payload": {
                "app_id": "e2e-audit-log-app",
                "collaborator_id": "collab-001",
                "user_id": "00000000-0000-0000-0000-000000000002",
                "user_email": "editor@example.com",
                "role": "editor"
            }
        }'::jsonb
    ),
    (
        'e2e000000000003',
        'e2e-portal',
        '2026-01-01 12:00:00',
        '',
        'site_admin.app.plan.updated',
        '127.0.0.1',
        'Mozilla/5.0',
        '',
        '{
            "id": "e2e000000000003",
            "seq": 3,
            "type": "site_admin.app.plan.updated",
            "context": {
                "app_id": "e2e-portal",
                "user_id": null,
                "timestamp": 1751374800,
                "triggered_by": "site_admin",
                "ip_address": "127.0.0.1",
                "user_agent": "Mozilla/5.0",
                "audit_context": {
                    "usage": "internal",
                    "actor_user_id": "00000000-0000-0000-0000-000000000001",
                    "http_referer": "",
                    "http_url": "http://localhost:4003/api/v1/apps/e2e-audit-log-app/plan"
                },
                "preferred_languages": []
            },
            "payload": {
                "app_id": "e2e-audit-log-app",
                "old_plan": "startups",
                "new_plan": "enterprise"
            }
        }'::jsonb
    )
ON CONFLICT DO NOTHING;
