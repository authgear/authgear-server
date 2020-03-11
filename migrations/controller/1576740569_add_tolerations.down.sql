ALTER TABLE microservice DROP COLUMN "tolerations";
ALTER TABLE microservice DROP COLUMN "builder_tolerations";

ALTER TABLE plan DROP COLUMN "microservice_tolerations";
ALTER TABLE plan DROP COLUMN "microservice_builder_tolerations";
