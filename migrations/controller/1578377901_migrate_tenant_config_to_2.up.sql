-- Put upgrade SQL here

-- Copy /app_config/hook to /hook
UPDATE config
SET config = jsonb_set(config, '{hook}', config #> '{app_config,hook}')
WHERE config #> '{app_config,hook}' IS NOT NULL AND jsonb_extract_path(config, 'version') = '"1"';

-- Remove /app_config/hook
UPDATE config
SET config = config #- '{app_config, hook}'
WHERE jsonb_extract_path(config, 'version') = '"1"';

-- Copy /app_config to /database_config
UPDATE config
SET config = jsonb_set(config, '{database_config}', config #> '{app_config}')
WHERE jsonb_extract_path(config, 'version') = '"1"';

-- No need to remove /app_config as it will replaced in the following step.
-- Copy /user_config to /app_config
UPDATE config
SET config = jsonb_set(config, '{app_config}', config #> '{user_config}')
WHERE jsonb_extract_path(config, 'version') = '"1"';

-- Remove /user_config
UPDATE config
SET config = config #- '{user_config}'
WHERE jsonb_extract_path(config, 'version') = '"1"';

-- Set /app_config/version to 2
UPDATE config
SET config = jsonb_set(config, '{app_config, version}', '"2"')
WHERE jsonb_extract_path(config, 'version') = '"1"';

-- Set /version to 2
UPDATE config
SET config = jsonb_set(config, '{version}', '"2"')
WHERE jsonb_extract_path(config, 'version') = '"1"';
