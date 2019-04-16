-- Put upgrade SQL here
ALTER TABLE cloud_code_route ALTER COLUMN backend_url DROP DEFAULT;
