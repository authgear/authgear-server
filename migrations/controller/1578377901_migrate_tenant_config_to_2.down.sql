-- Put downgrade SQL here

-- Remove /app_config/version
UPDATE config
SET config = config #- '{app_config,version}'
WHERE jsonb_extract_path(config, 'version') = '"2"';

-- Copy /app_config to /user_config
UPDATE config
SET config = jsonb_set(config, '{user_config}', config #> '{app_config}')
WHERE jsonb_extract_path(config, 'version') = '"2"';

-- No need to remove /app_config as it will replaced in the following step.
-- Copy /database_config to /app_config
UPDATE config
SET config = jsonb_set(config, '{app_config}', config #> '{database_config}')
WHERE jsonb_extract_path(config, 'version') = '"2"';

-- Remove /database_config
UPDATE config
SET config = config #- '{database_config}'
WHERE jsonb_extract_path(config, 'version') = '"2"';

-- Copy /hook to /app_config/hook
UPDATE config
SET config = jsonb_set(config, '{app_config, hook}', config #> '{hook}')
WHERE config #> '{hook}' IS NOT NULL AND jsonb_extract_path(config, 'version') = '"2"';

-- Remove /hook
UPDATE config
SET config = config #- '{hook}'
WHERE jsonb_extract_path(config, 'version') = '"2"';

-- Set /version to 1
UPDATE config
SET config = jsonb_set(config, '{version}', '"1"')
WHERE jsonb_extract_path(config, 'version') = '"2"';
