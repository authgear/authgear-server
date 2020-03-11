CREATE TABLE microservice_secret (
    microservice_id text NOT NULL REFERENCES microservice(id),
    secret_id text NOT NULL REFERENCES secret(id)
);
