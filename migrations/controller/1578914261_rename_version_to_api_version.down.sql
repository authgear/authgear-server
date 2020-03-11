-- Put downgrade SQL here

-- Set /app_config/version to 2
UPDATE config
SET config = jsonb_set(config, '{app_config,version}', '"2"')
WHERE jsonb_extract_path(config, 'api_version') = '"v2.1"';
-- Set /version to 2
UPDATE config
SET config = jsonb_set(config, '{version}', '"2"')
WHERE jsonb_extract_path(config, 'api_version') = '"v2.1"';
-- Remove /app_config/api_version
UPDATE config
SET config = config #- '{app_config,api_version}'
WHERE jsonb_extract_path(config, 'api_version') = '"v2.1"';
-- Remove /api_version
UPDATE config
SET config = config #- '{api_version}'
WHERE jsonb_extract_path(config, 'api_version') = '"v2.1"';
