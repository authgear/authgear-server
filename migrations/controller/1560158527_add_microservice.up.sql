-- Put upgrade SQL here
CREATE TABLE microservice (
  id uuid PRIMARY KEY,
  app_id uuid REFERENCES app(id) NOT NULL,
  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  created_by text REFERENCES _core_user(id) NOT NULL,

  name text NOT NULL,
  type text NOT NULL,
  path text NOT NULL,
  port integer NOT NULL,
  context text NOT NULL,
  dockerfile text NOT NULL,
  env jsonb NOT NULL,
  artifact_id uuid REFERENCES artifact(id) NOT NULL,

  image text,

  status text NOT NULL,

  k8s_resources jsonb NOT NULL
);

CREATE TABLE microservice_secret (
  microservice_id uuid REFERENCES microservice(id) NOT NULL,
  secret_id uuid REFERENCES secret(id) NOT NULL
);

CREATE TABLE deployment_microservice (
  deployment_id uuid REFERENCES deployment(id) NOT NULL,
  microservice_id uuid REFERENCES microservice(id) NOT NULL
);
