-- Fixed-date usage records for e2e-siteadmin-app-alpha.
-- Using 2020 dates avoids any dependency on NOW().

-- MAU: January 2020 = 100, February 2020 = 200
INSERT INTO _portal_usage_record (id, app_id, name, period, start_time, end_time, count)
VALUES
    (
        gen_random_uuid()::text,
        'e2e-siteadmin-app-alpha',
        'active-user',
        'monthly',
        '2020-01-01 00:00:00 UTC',
        '2020-02-01 00:00:00 UTC',
        100
    ),
    (
        gen_random_uuid()::text,
        'e2e-siteadmin-app-alpha',
        'active-user',
        'monthly',
        '2020-02-01 00:00:00 UTC',
        '2020-03-01 00:00:00 UTC',
        200
    ),
    -- Daily messaging records for 2020-01-10
    (
        gen_random_uuid()::text,
        'e2e-siteadmin-app-alpha',
        'sms-sent.north-america',
        'daily',
        '2020-01-10 00:00:00 UTC',
        '2020-01-11 00:00:00 UTC',
        10
    ),
    (
        gen_random_uuid()::text,
        'e2e-siteadmin-app-alpha',
        'sms-sent.other-regions',
        'daily',
        '2020-01-10 00:00:00 UTC',
        '2020-01-11 00:00:00 UTC',
        20
    ),
    (
        gen_random_uuid()::text,
        'e2e-siteadmin-app-alpha',
        'whatsapp-sent.north-america',
        'daily',
        '2020-01-10 00:00:00 UTC',
        '2020-01-11 00:00:00 UTC',
        30
    ),
    (
        gen_random_uuid()::text,
        'e2e-siteadmin-app-alpha',
        'whatsapp-sent.other-regions',
        'daily',
        '2020-01-10 00:00:00 UTC',
        '2020-01-11 00:00:00 UTC',
        40
    ),
    -- Daily messaging records for 2020-01-20
    (
        gen_random_uuid()::text,
        'e2e-siteadmin-app-alpha',
        'sms-sent.north-america',
        'daily',
        '2020-01-20 00:00:00 UTC',
        '2020-01-21 00:00:00 UTC',
        5
    ),
    (
        gen_random_uuid()::text,
        'e2e-siteadmin-app-alpha',
        'sms-sent.other-regions',
        'daily',
        '2020-01-20 00:00:00 UTC',
        '2020-01-21 00:00:00 UTC',
        5
    )
ON CONFLICT (app_id, name, period, start_time) DO UPDATE SET count = EXCLUDED.count;
