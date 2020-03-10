-- Put upgrade SQL here
CREATE TABLE deployment (
    id uuid PRIMARY KEY,
    app_id uuid REFERENCES app(id),
    created_at timestamp WITHOUT TIME ZONE NOT NULL,
    created_by text REFERENCES _core_user(id),
    version text NOT NULL,
    status text NOT NULL,
    UNIQUE (app_id, version)
);

CREATE TABLE deployment_cloud_code (
    deployment_id uuid REFERENCES deployment(id),
    cloud_code_id uuid REFERENCES cloud_code(id),
    CONSTRAINT deployment_cloud_code_pkey PRIMARY KEY (deployment_id, cloud_code_id)
);

ALTER TABLE app ADD COLUMN last_deployment_id UUID REFERENCES deployment(id);
