-- +migrate Up

ALTER TABLE _auth_resource ADD COLUMN access_policy jsonb NOT NULL DEFAULT '{}';
ALTER TABLE _auth_resource_scope ADD COLUMN access_policy jsonb NOT NULL DEFAULT '{}';

-- +migrate Down

ALTER TABLE _auth_resource_scope DROP COLUMN access_policy;
ALTER TABLE _auth_resource DROP COLUMN access_policy;
