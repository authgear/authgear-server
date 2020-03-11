
CREATE TABLE static_deployment (
  id text PRIMARY KEY,
  app_id text REFERENCES app(id) NOT NULL,
  deployment_id text REFERENCES deployment(id) NOT NULL,
  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  created_by text REFERENCES _core_user(id) NOT NULL,

  name text NOT NULL,
  path text NOT NULL,
  context text NOT NULL,
  fallback text NOT NULL,
  expires integer NOT NULL,
  raw_config jsonb NOT NULL,

  path_mapping jsonb NOT NULL,
  artifact_id text REFERENCES artifact(id) NOT NULL,
  status text NOT NULL
);
