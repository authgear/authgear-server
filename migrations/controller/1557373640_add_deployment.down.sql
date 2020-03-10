-- Put downgrade SQL here
ALTER TABLE app DROP COLUMN "last_deployment_id";
DROP TABLE deployment_cloud_code;
DROP TABLE deployment;
