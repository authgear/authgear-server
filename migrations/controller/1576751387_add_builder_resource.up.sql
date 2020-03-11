ALTER TABLE plan ADD COLUMN "microservice_builder_resources" jsonb;

UPDATE plan SET "microservice_builder_resources" = '
{
    "requests": {
        "cpu": "750m",
        "memory": "2Gi"
    },
    "limits": {
        "cpu": "1",
        "memory": "4Gi"
    }
}';

ALTER TABLE plan ALTER COLUMN "microservice_builder_resources" SET NOT NULL;

ALTER TABLE microservice ADD COLUMN "builder_resources" jsonb;
UPDATE microservice SET "builder_resources" = '{}';
ALTER TABLE microservice ALTER COLUMN "builder_resources" SET NOT NULL;
