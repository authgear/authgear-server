-- Put downgrade SQL here
ALTER TABLE deployment_route
    ADD COLUMN "target_path" text,
    ADD COLUMN "backend_url" text;

UPDATE deployment_route SET "target_path" = json_extract_path_text(("type_config")::JSON, 'target_path');
UPDATE deployment_route SET "backend_url" = json_extract_path_text(("type_config")::JSON, 'backend_url');

ALTER TABLE deployment_route
    ALTER COLUMN "type" SET NOT NULL,
    ALTER COLUMN "backend_url" SET NOT NULL;

ALTER TABLE deployment_route
    DROP COLUMN "type",
    DROP COLUMN "type_config",
    DROP COLUMN "is_last_deployment";

ALTER TABLE deployment_route RENAME COLUMN deployment_version TO version;
ALTER TABLE deployment_route RENAME TO cloud_code_route;

UPDATE cloud_code_route SET version = '';
