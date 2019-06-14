-- Put upgrade SQL here
ALTER TABLE cloud_code_route RENAME TO deployment_route;
ALTER TABLE deployment_route RENAME COLUMN version TO deployment_version;
ALTER TABLE deployment_route
    ADD COLUMN "type" text,
    ADD COLUMN "type_config" JSONB NOT NULL DEFAULT '{}'::JSONB,
    ADD COLUMN "is_last_deployment" BOOLEAN NOT NULL DEFAULT false;

UPDATE deployment_route SET type = 'http-handler';
UPDATE deployment_route SET type_config = jsonb_build_object(
	'target_path', target_path,
	'backend_url', backend_url
);
UPDATE deployment_route SET is_last_deployment = true;

ALTER TABLE deployment_route
    ALTER COLUMN "type" SET NOT NULL;

ALTER TABLE deployment_route
    DROP COLUMN "target_path",
    DROP COLUMN "backend_url";

-- Update deployment version with last deployment
UPDATE
	deployment_route SET deployment_version = sub.v
FROM (
 	SELECT
 		COALESCE(deployment.version, '') as v,
 		app.id as aid
 	FROM
 	    deployment
 	RIGHT JOIN
 		app
 	ON app.last_deployment_id = deployment.id
) as sub
WHERE
    sub.aid = deployment_route.app_id;
