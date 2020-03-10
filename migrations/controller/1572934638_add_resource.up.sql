ALTER TABLE plan ADD COLUMN "microservice_default_resources" jsonb;
ALTER TABLE plan ADD COLUMN "resource_quota" jsonb;

UPDATE plan SET "microservice_default_resources" = '
{
    "requests": {
        "cpu": "125m",
        "memory": "128Mi"
    },
    "limits": {
        "cpu": "250m",
        "memory": "256Mi"
    }
}';

UPDATE plan SET "resource_quota" = '
{
    "requests": {
        "cpu": "250m",
        "memory": "256Mi"
    },
    "limits": {
        "cpu": "500m",
        "memory": "512Mi"
    }
}';

ALTER TABLE plan ALTER COLUMN "microservice_default_resources" SET NOT NULL;
ALTER TABLE plan ALTER COLUMN "resource_quota" SET NOT NULL;

ALTER TABLE microservice ADD COLUMN "resources" jsonb;
UPDATE microservice SET "resources" = '{}';
ALTER TABLE microservice ALTER COLUMN "resources" SET NOT NULL;
