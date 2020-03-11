CREATE TABLE cloud_code_backend (
    id text PRIMARY KEY,
    type text NOT NULL,
    controller_url text NOT NULL,
    gateway_url text NOT NULL,
    is_default boolean NOT NULL DEFAULT false
);

ALTER TABLE cloud_code_backend ADD CONSTRAINT cloud_code_backend_type CHECK ("type" = 'fission');

CREATE TABLE cloud_code (
    created_at timestamp without time zone NOT NULL,
    app_id text NOT NULL REFERENCES app(id),
    name text NOT NULL,
    type text NOT NULL,
    trigger_type text NOT NULL,
    trigger_config jsonb NOT NULL,
    environment text NOT NULL,
    version text NOT NULL,
    config jsonb NOT NULL,
    status text NOT NULL,
    artifact_id text NOT NULL REFERENCES artifact(id),
    backend_id text NOT NULL REFERENCES cloud_code_backend(id),
    backend_resources jsonb NOT NULL,
    entry_point text NOT NULL,
    created_by text REFERENCES _core_user(id),
    id text PRIMARY KEY UNIQUE
);

ALTER TABLE cloud_code ADD CONSTRAINT cloud_code_type CHECK ("type" = 'http-handler');
ALTER TABLE cloud_code ADD CONSTRAINT cloud_code_trigger_type CHECK ("trigger_type" = 'http');

CREATE TABLE deployment_cloud_code (
    deployment_id text REFERENCES deployment(id),
    cloud_code_id text REFERENCES cloud_code(id),
    CONSTRAINT deployment_cloud_code_pkey PRIMARY KEY (deployment_id, cloud_code_id)
);

CREATE TABLE cloud_code_secret (
    cloud_code_id text REFERENCES cloud_code(id),
    secret_id text REFERENCES secret(id),
    CONSTRAINT cloud_code_secret_pkey PRIMARY KEY (cloud_code_id, secret_id)
);

