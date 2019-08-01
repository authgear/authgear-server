CREATE TABLE deployment_hook(
    id uuid PRIMARY KEY,
    created_at timestamp without time zone NOT NULL,
    deployment_version text NOT NULL,
    hooks jsonb NOT NULL,
    app_id uuid NOT NULL REFERENCES app(id),
    is_last_deployment boolean NOT NULL DEFAULT false
);
