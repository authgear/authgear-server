-- +migrate Up
CREATE TABLE _auth_identity_ldap
(
    id                 text  PRIMARY KEY REFERENCES _auth_identity (id),
    app_id             text  NOT NULL,
    server_name        text  NOT NULL,
    user_id_attribute  text  NOT NULL,
    user_id_value      text  NOT NULL,
    claims             jsonb NOT NULL,
    raw_entry_json     jsonb NOT NULL
);
ALTER TABLE _auth_identity_ldap
    ADD CONSTRAINT _auth_identity_ldap_unique UNIQUE (app_id, server_name, user_id_attribute, user_id_value);

-- +migrate Down
DROP TABLE _auth_identity_ldap;
DELETE FROM _auth_identity WHERE "type" = 'ldap';
