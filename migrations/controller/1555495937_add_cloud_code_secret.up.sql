-- Put upgrade SQL here
CREATE TABLE cloud_code_secret (
    cloud_code_id uuid REFERENCES cloud_code(id),
    secret_id uuid REFERENCES secret(id),
    CONSTRAINT cloud_code_secret_pkey PRIMARY KEY (cloud_code_id, secret_id)
);
