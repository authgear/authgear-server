-- +migrate Up

CREATE TABLE _portal_plan
(
    id                 text     PRIMARY KEY,
    name               text     NOT NULL,
    feature_config     jsonb    NOT NULL,
    created_at         timestamp WITHOUT TIME ZONE NOT NULL,
    updated_at         timestamp WITHOUT TIME ZONE NOT NULL,
    UNIQUE (name)
);

ALTER TABLE _portal_config_source ADD COLUMN plan_name text;
UPDATE _portal_config_source SET plan_name = '';
ALTER TABLE _portal_config_source ALTER COLUMN plan_name SET NOT NULL;

-- +migrate Down

ALTER TABLE _portal_config_source DROP COLUMN plan_name;

DROP TABLE _portal_plan;
