CREATE TABLE root_domain (
	id text PRIMARY KEY,
    created_by text REFERENCES _core_user(id),
	created_at timestamp WITHOUT TIME ZONE NOT NULL,
    verified_by text REFERENCES _core_user(id),
	verified_at timestamp WITHOUT TIME ZONE,
	domain text NOT NULL,
	app_id text REFERENCES app(id) NOT NULL,
    verified boolean NOT NULL DEFAULT FALSE,
    dns_records jsonb NOT NULL,
    UNIQUE (app_id, domain)
);

CREATE TABLE custom_domain (
	id text PRIMARY KEY,
    created_by text REFERENCES _core_user(id),
	created_at timestamp WITHOUT TIME ZONE NOT NULL,
    verified_by text REFERENCES _core_user(id),
	verified_at timestamp WITHOUT TIME ZONE,
	domain text NOT NULL,
	app_id text REFERENCES app(id) NOT NULL,
    verified boolean NOT NULL DEFAULT FALSE,
    dns_records jsonb NOT NULL,
	redirect_domain text NOT NULL,
	tls_secret_id text REFERENCES secret(id),
	tls_secret_expiry timestamp WITHOUT TIME ZONE,
    root_domain_id text REFERENCES root_domain(id) NOT NULL,
    UNIQUE (app_id, domain)
);
