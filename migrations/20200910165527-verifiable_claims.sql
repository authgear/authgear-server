-- +migrate Up
CREATE TABLE _auth_verified_claim
(
    id         text PRIMARY KEY,
    app_id     text                        NOT NULL,
    user_id    text                        NOT NULL REFERENCES _auth_user (id),
    name       text                        NOT NULL,
    value      text                        NOT NULL,
    created_at timestamp without time zone NOT NULL
);

-- Remove authenticators without primary/secondary tags
DELETE FROM _auth_authenticator_oob ao USING _auth_authenticator a WHERE a.id = ao.id AND NOT (a.tag ? 'authentication:primary_authenticator' OR a.tag ? 'authentication:secondary_authenticator');
DELETE FROM _auth_authenticator_password ap USING _auth_authenticator a WHERE a.id = ap.id AND NOT (a.tag ? 'authentication:primary_authenticator' OR a.tag ? 'authentication:secondary_authenticator');
DELETE FROM _auth_authenticator_totp at USING _auth_authenticator a WHERE a.id = at.id AND NOT (a.tag ? 'authentication:primary_authenticator' OR a.tag ? 'authentication:secondary_authenticator');
DELETE FROM _auth_authenticator a WHERE NOT (a.tag ? 'authentication:primary_authenticator' OR a.tag ? 'authentication:secondary_authenticator');

-- +migrate Down
DROP TABLE _auth_verified_claim;
