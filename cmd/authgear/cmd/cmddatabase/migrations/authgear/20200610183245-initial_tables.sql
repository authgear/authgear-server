-- +migrate Up

CREATE TABLE _auth_user
(
    id            text PRIMARY KEY,
    app_id        text                        NOT NULL,
    created_at    timestamp without time zone NOT NULL,
    updated_at    timestamp without time zone NOT NULL,
    last_login_at timestamp without time zone,
    metadata      jsonb
);

CREATE SEQUENCE _auth_event_sequence;

CREATE TABLE _auth_password_history
(
    id         text PRIMARY KEY,
    app_id     text                        NOT NULL,
    created_at timestamp without time zone NOT NULL,
    user_id    text                        NOT NULL REFERENCES _auth_user (id),
    password   text                        NOT NULL
);

CREATE TABLE _auth_oauth_authorization
(
    id         text PRIMARY KEY,
    app_id     text                        NOT NULL,
    client_id  text                        NOT NULL,
    user_id    text                        NOT NULL REFERENCES _auth_user (id),
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    scopes     jsonb                       NOT NULL
);
ALTER TABLE _auth_oauth_authorization
    ADD CONSTRAINT _auth_oauth_authorization_key UNIQUE (app_id, user_id, client_id);

CREATE TABLE _auth_identity
(
    id      text PRIMARY KEY,
    app_id  text NOT NULL,
    type    text NOT NULL,
    user_id text NOT NULL REFERENCES _auth_user (id)
);

CREATE TABLE _auth_identity_anonymous
(
    id     text PRIMARY KEY REFERENCES _auth_identity (id),
    app_id text  NOT NULL,
    key_id text  NOT NULL,
    key    jsonb NOT NULL
);
ALTER TABLE _auth_identity_anonymous
    ADD CONSTRAINT _auth_identity_anonymous_key UNIQUE (app_id, key_id);

CREATE TABLE _auth_identity_login_id
(
    id                text PRIMARY KEY REFERENCES _auth_identity (id),
    app_id            text  NOT NULL,
    login_id_key      text  NOT NULL,
    login_id          text  NOT NULL,
    claims            jsonb NOT NULL,
    original_login_id text  NOT NULL,
    unique_key        text  NOT NULL
);
ALTER TABLE _auth_identity_login_id
    ADD CONSTRAINT _auth_identity_login_id_key UNIQUE (app_id, unique_key);

CREATE TABLE _auth_identity_oauth
(
    id               text PRIMARY KEY REFERENCES _auth_identity (id),
    app_id           text                        NOT NULL,
    created_at       timestamp without time zone NOT NULL,
    updated_at       timestamp without time zone NOT NULL,
    provider_type    text                        NOT NULL,
    provider_keys    jsonb                       NOT NULL DEFAULT '{}'::jsonb,
    provider_user_id text                        NOT NULL,
    claims           jsonb                       NOT NULL,
    profile          jsonb
);
ALTER TABLE _auth_identity_oauth
    ADD CONSTRAINT _auth_identity_oauth_key UNIQUE (app_id, provider_type, provider_keys, provider_user_id);

CREATE TABLE _auth_authenticator
(
    id      text PRIMARY KEY,
    app_id  text NOT NULL,
    type    text NOT NULL,
    user_id text NOT NULL REFERENCES _auth_user (id)
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

CREATE TABLE _auth_authenticator_oob
(
    id         text PRIMARY KEY REFERENCES _auth_authenticator (id),
    app_id     text                        NOT NULL,
    created_at timestamp without time zone NOT NULL,
    channel    text                        NOT NULL,
    phone      text                        NOT NULL,
    email      text                        NOT NULL
);

CREATE TABLE _auth_authenticator_password
(
    id            text PRIMARY KEY REFERENCES _auth_authenticator (id),
    app_id        text NOT NULL,
    password_hash text NOT NULL
);

CREATE TABLE _auth_authenticator_recovery_code
(
    id         text PRIMARY KEY REFERENCES _auth_authenticator (id),
    app_id     text                        NOT NULL,
    code       text                        NOT NULL,
    created_at timestamp without time zone NOT NULL,
    consumed   boolean                     NOT NULL
);

CREATE TABLE _auth_authenticator_totp
(
    id           text PRIMARY KEY REFERENCES _auth_authenticator (id),
    app_id       text                        NOT NULL,
    created_at   timestamp without time zone NOT NULL,
    secret       text                        NOT NULL,
    display_name text                        NOT NULL
);

-- +migrate Down
DROP TABLE _auth_authenticator_totp;
DROP TABLE _auth_authenticator_recovery_code;
DROP TABLE _auth_authenticator_password;
DROP TABLE _auth_authenticator_oob;
DROP TABLE _auth_authenticator_bearer_token;
DROP TABLE _auth_authenticator;
DROP TABLE _auth_identity_oauth;
DROP TABLE _auth_identity_login_id;
DROP TABLE _auth_identity_anonymous;
DROP TABLE _auth_identity;
DROP TABLE _auth_oauth_authorization;
DROP TABLE _auth_password_history;
DROP SEQUENCE _auth_event_sequence;
DROP TABLE _auth_user;
