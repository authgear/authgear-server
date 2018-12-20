-- Put upgrade SQL here
CREATE TABLE _auth_verify_code (
    id text PRIMARY KEY,
    user_id text NOT NULL REFERENCES _core_user(id),
    record_key text NOT NULL,
    record_value text NOT NULL,
    code text NOT NULL,
    consumed boolean NOT NULL DEFAULT false,
    created_at timestamp without time zone NOT NULL
);
