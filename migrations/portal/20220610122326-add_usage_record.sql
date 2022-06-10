-- +migrate Up

CREATE TABLE _portal_usage_record (
    id               text PRIMARY KEY,
    app_id           text                        NOT NULL,
    name             text                        NOT NULL,
    period           text                        NOT NULL,
    start_time       timestamp WITHOUT TIME ZONE NOT NULL,
    end_time         timestamp WITHOUT TIME ZONE NOT NULL,
    count            integer                     NOT NULL,
    alert_data       jsonb,
    stripe_timestamp timestamp WITHOUT TIME ZONE,
    UNIQUE (app_id, name, period, start_time)
);

-- +migrate Down
DROP TABLE _portal_usage_record;
