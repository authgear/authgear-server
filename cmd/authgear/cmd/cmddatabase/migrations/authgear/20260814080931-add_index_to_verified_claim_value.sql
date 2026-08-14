-- +migrate Up
CREATE INDEX _auth_verified_claim_app_id_name_value ON _auth_verified_claim (app_id, name, value);

-- +migrate Down
DROP INDEX _auth_verified_claim_app_id_name_value;
