-- +migrate Up
CREATE TABLE _audit_analytic_count
(
    id         text PRIMARY KEY,
    app_id     text NOT NULL,
    type       text NOT NULL,
    count      int  NOT NULL,
    date       date NOT NULL,
    UNIQUE (app_id, type, date)
);

-- +migrate Down

DROP TABLE _audit_analytic_count;
