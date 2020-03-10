-- Put upgrade SQL here
CREATE TABLE secret (
    id uuid PRIMARY KEY,
    app_id uuid REFERENCES app(id),
    name text NOT NULL,
    k8s_secret_name text NOT NULL,
    created_at timestamp WITHOUT TIME ZONE NOT NULL,
    created_by text REFERENCES _core_user(id),
    updated_at timestamp WITHOUT TIME ZONE NOT NULL,
    updated_by text REFERENCES _core_user(id),
    deleted bool NOT NULL
);
