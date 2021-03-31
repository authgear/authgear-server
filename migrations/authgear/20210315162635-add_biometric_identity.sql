-- +migrate Up
CREATE TABLE _auth_identity_biometric
(
    id          text PRIMARY KEY REFERENCES _auth_identity (id),
    app_id      text  NOT NULL,
    key_id      text  NOT NULL,
    key         jsonb NOT NULL,
    device_info jsonb NOT NULL
);
ALTER TABLE _auth_identity_biometric
    ADD CONSTRAINT _auth_identity_biometric_key UNIQUE (app_id, key_id);

-- +migrate Down
DROP TABLE _auth_identity_biometric;
DELETE FROM _auth_identity WHERE "type" = 'biometric';
