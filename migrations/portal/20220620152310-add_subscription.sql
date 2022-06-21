-- +migrate Up
CREATE TABLE _portal_subscription (
    id                          text PRIMARY KEY,
    app_id                      text NOT NULL,
    stripe_checkout_session_id  text NOT NULL,
    stripe_customer_id          text NOT NULL,
    stripe_subscription_id      text NOT NULL,
    UNIQUE (app_id)
);

-- +migrate Down
DROP TABLE _portal_subscription;
