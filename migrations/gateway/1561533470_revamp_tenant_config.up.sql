-- Put upgrade SQL here
ALTER TABLE config ADD COLUMN "config_new" JSONB;
-- Minimal migration on essential fields
UPDATE config SET "config_new" = sub.new_config
FROM (
  SELECT id, (
    jsonb_build_object(
      'version', '1',
      'app_name', c.config #> '{APP_NAME}',
      'app_config', jsonb_build_object(
        'database_url', c.config #> '{DATABASE_URL}'
      ),
      'user_config', jsonb_build_object(
        'api_key', c.config #> '{API_KEY}',
        'master_key', c.config #> '{MASTER_KEY}',
        'auth', jsonb_build_object(
          'login_id_keys', c.config #> '{AUTH,LOGIN_IDS_KEY_WHITELIST}'
        ),
        'token_store', jsonb_build_object(
          'secret', c.config #> '{TOKEN_STORE,SECRET}'
        )
      )
    )
  ) AS new_config
  FROM config c
) AS sub
WHERE config.id = sub.id;
ALTER TABLE config ALTER COLUMN "config_new" SET NOT NULL;
ALTER TABLE config RENAME COLUMN "config" TO "config_old";
ALTER TABLE config RENAME COLUMN "config_new" TO "config";
ALTER TABLE config ALTER COLUMN "config_old" SET DEFAULT '{}'::JSONB;
