-- +migrate Up
CREATE TABLE _portal_tutorial_progress
(
    app_id text PRIMARY KEY,
    data jsonb NOT NULL
);

-- +migrate Down
DROP TABLE _portal_tutorial_progress;
