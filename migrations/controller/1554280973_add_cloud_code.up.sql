-- Put upgrade SQL here
CREATE TABLE cloud_code_backend (
    id uuid PRIMARY KEY,
    type text NOT NULL,
    controller_url text NOT NULL,
    gateway_url text NOT NULL,
    is_default boolean NOT NULL DEFAULT false
);

ALTER TABLE cloud_code_backend ADD CONSTRAINT cloud_code_backend_type CHECK ("type" = 'fission');

CREATE TABLE cloud_code (
    id uuid PRIMARY KEY,
    created_at timestamp WITHOUT TIME ZONE NOT NULL,
    created_by uuid REFERENCES app(id),
    app_id uuid REFERENCES app(id) NOT NULL,
    name text NOT NULL,
    type text NOT NULL,
    trigger_type text NOT NULL,
    trigger_config jsonb NOT NULL,
    environment text NOT NULL,
    version text NOT NULL,
    config jsonb NOT NULL,
    status text NOT NULL,
    artifact_id uuid REFERENCES artifact(id) NOT NULL,
    backend_id uuid REFERENCES cloud_code_backend(id) NOT NULL,
    backend_resources jsonb NOT NULL
);

ALTER TABLE cloud_code ADD CONSTRAINT cloud_code_type CHECK ("type" = 'http-handler');
ALTER TABLE cloud_code ADD CONSTRAINT cloud_code_trigger_type CHECK ("trigger_type" = 'http');
