-- +migrate Up
ALTER TABLE _portal_subscription ADD COLUMN cancelled_at timestamp without time zone;
ALTER TABLE _portal_subscription ADD COLUMN ended_at timestamp without time zone;

CREATE TABLE _portal_historical_subscription (
    id                          text PRIMARY KEY,
    app_id                      text NOT NULL,
    stripe_customer_id          text NOT NULL,
    stripe_subscription_id      text NOT NULL,
    subscription_created_at     timestamp WITHOUT TIME ZONE NOT NULL,
    subscription_updated_at     timestamp WITHOUT TIME ZONE NOT NULL,
    subscription_cancelled_at   timestamp WITHOUT TIME ZONE,
    subscription_ended_at       timestamp WITHOUT TIME ZONE,
    created_at                  timestamp WITHOUT TIME ZONE NOT NULL
);

-- +migrate Down
DROP TABLE _portal_historical_subscription;

ALTER TABLE _portal_subscription DROP COLUMN cancelled_at;
ALTER TABLE _portal_subscription DROP COLUMN ended_at;
