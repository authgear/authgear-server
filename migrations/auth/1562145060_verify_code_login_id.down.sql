ALTER TABLE _auth_verify_code RENAME COLUMN login_id_key TO record_key;
ALTER TABLE _auth_verify_code RENAME COLUMN login_id TO record_value;
