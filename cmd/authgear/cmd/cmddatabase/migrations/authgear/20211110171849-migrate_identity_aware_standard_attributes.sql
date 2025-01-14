-- +migrate Up
UPDATE _auth_user
SET standard_attributes = phase2.standard_attributes
-- phase2 is a table looks like
-- USER_A, { "email": "user_a@example.com", "phone_number": "+85298765432" }
-- USER_B, { "email": "user_b@example.com" }
FROM (
  SELECT
    user_id,
    jsonb_object_agg(key, value) as standard_attributes
  -- phase1 is a table looks like
  -- USER_A, 'email', 'user_a@example.com', CREATED_AT
  -- USER_A, 'phone_number', '+85298765432', CREATED_AT
  -- USER_B, 'email', 'user_b@example.com', CREATED_AT
  FROM (
    SELECT
    user_id,
    (
      CASE
      WHEN claims ? 'email' THEN 'email'
      WHEN claims ? 'phone_number' THEN 'phone_number'
      WHEN claims ? 'preferred_username' THEN 'preferred_username'
      END
    ) as key,
    (
      CASE
      WHEN claims ? 'email' THEN claims ->> 'email'
      WHEN claims ? 'phone_number' THEN claims ->> 'phone_number'
      WHEN claims ? 'preferred_username' THEN claims ->> 'preferred_username'
      END
    ) AS value,
    created_at
    -- phase0 is a table looks like
    -- USER_A, { "email": "user_a@example.com" }, CREATED_AT
    -- USER_A, { "phone_number": "+85298765432" }, CREATED_AT
    -- USER_B, { "email": "user_b@example.com" }, CREATED_AT
    FROM (
      SELECT
      t1.user_id AS user_id,
      CASE
        WHEN t1.type = 'login_id' THEN t2.claims
        WHEN t1.type = 'oauth' THEN t3.claims
      END AS claims,
      t1.created_at AS created_at
      FROM _auth_identity AS t1
      LEFT OUTER JOIN _auth_identity_login_id AS t2 ON t1.id = t2.id
      LEFT OUTER JOIN _auth_identity_oauth AS t3 on t1.id = t3.id
      -- This order is very important here because the result of jsonb_object_agg is order-sensitive.
      ORDER BY created_at DESC
    ) AS phase0
  ) AS phase1
  WHERE key IS NOT NULL AND value IS NOT NULL
  GROUP BY user_id
) AS phase2
WHERE phase2.user_id = id;

-- +migrate Down

-- No thing needed to be migrated down.
