-- +migrate Up
CREATE TABLE _auth_identity_passkey
(
    id                   text  PRIMARY KEY REFERENCES _auth_identity (id),
    app_id               text  NOT NULL,
    credential_id        text  NOT NULL,
    creation_options     jsonb NOT NULL,
    attestation_response jsonb NOT NULL
);
ALTER TABLE _auth_identity_passkey
    ADD CONSTRAINT _auth_identity_passkey_credential_id UNIQUE (credential_id);

CREATE TABLE _auth_authenticator_passkey
(
    id                   text   PRIMARY KEY REFERENCES _auth_authenticator (id),
    app_id               text   NOT NULL,
    credential_id        text   NOT NULL,
    creation_options     jsonb  NOT NULL,
    attestation_response jsonb  NOT NULL,
    sign_count           bigint NOT NULL
);
ALTER TABLE _auth_authenticator_passkey
    ADD CONSTRAINT _auth_authenticator_passkey_credential_id UNIQUE (credential_id);

-- +migrate Down
DROP TABLE _auth_identity_passkey;
DELETE FROM _auth_identity WHERE "type" = 'passkey';

DROP TABLE _auth_authenticator_passkey;
DELETE FROM _auth_authenticator WHERE "type" = 'passkey';
