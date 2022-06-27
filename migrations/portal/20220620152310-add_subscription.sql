-- +migrate Up
CREATE TABLE _portal_subscription (
    id                          text PRIMARY KEY,
    app_id                      text NOT NULL,
    stripe_customer_id          text NOT NULL,
    stripe_subscription_id      text NOT NULL,
    created_at                  timestamp WITHOUT TIME ZONE NOT NULL,
    updated_at                  timestamp WITHOUT TIME ZONE NOT NULL,
    UNIQUE (app_id)
);

CREATE TABLE _portal_subscription_checkout (
    id                          text PRIMARY KEY,
    app_id                      text NOT NULL,
    stripe_checkout_session_id  text NOT NULL,
    stripe_customer_id          text,
    status                      text NOT NULL,
    created_at                  timestamp WITHOUT TIME ZONE NOT NULL,
    updated_at                  timestamp WITHOUT TIME ZONE NOT NULL,
    expire_at                   timestamp WITHOUT TIME ZONE NOT NULL,
    UNIQUE (stripe_checkout_session_id),
    UNIQUE (stripe_customer_id)
);

-- +migrate Down
DROP TABLE _portal_subscription;

DROP TABLE _portal_subscription_checkout;
