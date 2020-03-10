ALTER TABLE plan ADD COLUMN "microservice_tolerations" jsonb;
ALTER TABLE plan ADD COLUMN "microservice_builder_tolerations" jsonb;

UPDATE plan SET "microservice_tolerations" = '
[{
    "key": "preemptible",
    "operator": "Equal",
    "value": "true",
    "effect": "NoSchedule"
}]';

UPDATE plan SET "microservice_builder_tolerations" = '
[{
    "key": "preemptible",
    "operator": "Equal",
    "value": "true",
    "effect": "NoSchedule"
}]';

ALTER TABLE plan ALTER COLUMN "microservice_tolerations" SET NOT NULL;
ALTER TABLE plan ALTER COLUMN "microservice_builder_tolerations" SET NOT NULL;

ALTER TABLE microservice ADD COLUMN "tolerations" jsonb;
UPDATE microservice SET "tolerations" = '[]';
ALTER TABLE microservice ALTER COLUMN "tolerations" SET NOT NULL;

ALTER TABLE microservice ADD COLUMN "builder_tolerations" jsonb;
UPDATE microservice SET "builder_tolerations" = '[]';
ALTER TABLE microservice ALTER COLUMN "builder_tolerations" SET NOT NULL;
