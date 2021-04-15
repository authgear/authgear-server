-- +migrate Up

CREATE TABLE _portal_config_source
(
    id                 text PRIMARY KEY,
    app_id             text                        NOT NULL,
    created_at         timestamp WITHOUT TIME ZONE NOT NULL,
    updated_at         timestamp WITHOUT TIME ZONE NOT NULL,
    data               jsonb                       NOT NULL,
    UNIQUE (app_id)
);

-- +migrate Down

DROP TABLE _portal_config_source;
