-- Put upgrade SQL here

-- Set /app_config/api_version to v2.1
UPDATE config
SET config = jsonb_set(config, '{app_config,api_version}', '"v2.1"')
WHERE jsonb_extract_path(config, 'version') = '"2"';
-- Set /api_version to v2.1
UPDATE config
SET config = jsonb_set(config, '{api_version}', '"v2.1"')
WHERE jsonb_extract_path(config, 'version') = '"2"';
-- Remove /app_config/version
UPDATE config
SET config = config #- '{app_config,version}'
WHERE jsonb_extract_path(config, 'version') = '"2"';
-- Remove /version
UPDATE config
SET config = config #- '{version}'
WHERE jsonb_extract_path(config, 'version') = '"2"';
