-- +migrate Up
CREATE TABLE _portal_user_app_quota
(
    user_id      text PRIMARY KEY,
    max_own_apps integer NOT NULL,
    UNIQUE (user_id)
);

-- +migrate Down

DROP TABLE _portal_user_app_quota;
