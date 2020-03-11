ALTER TABLE microservice DROP COLUMN "resources";

ALTER TABLE plan DROP COLUMN "resource_quota";
ALTER TABLE plan DROP COLUMN "microservice_default_resources";
