-- +migrate Up
CREATE TABLE _portal_pending_domain
(
    id                 text PRIMARY KEY,
    app_id             text                        NOT NULL,
    created_at         timestamp WITHOUT TIME ZONE NOT NULL,
    domain             text                        NOT NULL,
    apex_domain        text                        NOT NULL,
    verification_nonce text                        NOT NULL,
    UNIQUE (app_id, apex_domain)
);

CREATE TABLE _portal_domain
(
    id                 text PRIMARY KEY,
    app_id             text                        NOT NULL,
    created_at         timestamp WITHOUT TIME ZONE NOT NULL,
    domain             text                        NOT NULL,
    apex_domain        text                        NOT NULL,
    verification_nonce text                        NOT NULL,
    UNIQUE (apex_domain)
);

-- +migrate Down

DROP TABLE _portal_pending_domain;
DROP TABLE _portal_domain;
