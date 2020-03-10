-- Put upgrade SQL here
CREATE TABLE artifact (
    id uuid PRIMARY KEY,
    app_id uuid REFERENCES app(id),
    created_at timestamp WITHOUT TIME ZONE NOT NULL,
    created_by text REFERENCES _core_user(id),
    checksum text NOT NULL,
    storage_type text NOT NULL,
    storage_data jsonb NOT NULL
);
