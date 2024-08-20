-- +migrate Up
CREATE TABLE _auth_identity_ldap
(
    id                        text  PRIMARY KEY REFERENCES _auth_identity (id),
    app_id                    text  NOT NULL,
    server_name               text  NOT NULL,
    user_id_attribute_name    text  NOT NULL,
    user_id_attribute_value   bytea NOT NULL,
    claims                    jsonb NOT NULL,
    raw_entry_json            jsonb NOT NULL
);
ALTER TABLE _auth_identity_ldap
    ADD CONSTRAINT _auth_identity_ldap_unique UNIQUE (app_id, server_name, user_id_attribute_name, user_id_attribute_value);

CREATE INDEX _auth_identity_ldap_claim_preferred_username ON _auth_identity_ldap (app_id, (claims ->> 'preferred_username'));
CREATE INDEX _auth_identity_ldap_claim_phone_number ON _auth_identity_ldap (app_id, (claims ->> 'phone_number'));
CREATE INDEX _auth_identity_ldap_claim_email ON _auth_identity_ldap (app_id, (claims ->> 'email'));

-- +migrate Down
DROP TABLE _auth_identity_ldap;
DELETE FROM _auth_identity WHERE "type" = 'ldap';
