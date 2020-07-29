-- +migrate Up
DROP TABLE _auth_authenticator_bearer_token;
DROP TABLE _auth_authenticator_recovery_code;

CREATE TABLE _auth_recovery_code
(
    id         text PRIMARY KEY,
    app_id     text                        NOT NULL,
    user_id    text                        NOT NULL REFERENCES _auth_user (id),
    code       text                        NOT NULL,
    created_at timestamp without time zone NOT NULL,
    consumed   boolean                     NOT NULL
);

ALTER TABLE _auth_authenticator_oob DROP COLUMN identity_id;

-- +migrate Down

ALTER TABLE _auth_authenticator_oob ADD COLUMN identity_id text references _auth_identity;

DROP TABLE _auth_recovery_code;

CREATE TABLE _auth_authenticator_recovery_code
(
    id         text PRIMARY KEY REFERENCES _auth_authenticator (id),
    app_id     text                        NOT NULL,
    code       text                        NOT NULL,
    created_at timestamp without time zone NOT NULL,
    consumed   boolean                     NOT NULL
);

CREATE TABLE _auth_authenticator_bearer_token
(
    id         text PRIMARY KEY REFERENCES _auth_authenticator (id),
    app_id     text                        NOT NULL,
    parent_id  text                        NOT NULL REFERENCES _auth_authenticator (id),
    token      text                        NOT NULL,
    created_at timestamp without time zone NOT NULL,
    expire_at  timestamp without time zone NOT NULL
);