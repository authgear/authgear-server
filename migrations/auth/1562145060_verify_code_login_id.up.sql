ALTER TABLE _auth_verify_code RENAME COLUMN record_key TO login_id_key;
ALTER TABLE _auth_verify_code RENAME COLUMN record_value TO login_id;
