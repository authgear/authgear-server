-- +migrate Up
CREATE INDEX _auth_identity_login_id_claim_preferred_username ON _auth_identity_login_id (app_id, (claims ->> 'preferred_username'));

-- +migrate Down
DROP INDEX _auth_identity_login_id_claim_preferred_username;
