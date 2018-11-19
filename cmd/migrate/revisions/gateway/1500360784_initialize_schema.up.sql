CREATE TABLE config (
	id uuid PRIMARY KEY,
	created_at timestamp WITHOUT TIME ZONE NOT NULL,
	updated_at timestamp WITHOUT TIME ZONE NOT NULL,
	config jsonb NOT NULL,
	app_id uuid NOT NULL
);

CREATE TABLE plan (
	id uuid PRIMARY KEY,
	created_at timestamp WITHOUT TIME ZONE NOT NULL,
	updated_at timestamp WITHOUT TIME ZONE NOT NULL,
	name text NOT NULL,
	auth_enabled boolean NOT NULL DEFAULT FALSE
	);

CREATE TABLE app (
	id uuid PRIMARY KEY,
	created_at timestamp WITHOUT TIME ZONE NOT NULL,
	updated_at timestamp WITHOUT TIME ZONE NOT NULL,
	name text NOT NULL,
	plan_id uuid REFERENCES plan(id) NOT NULL,
	config_id uuid REFERENCES config(id) NOT NULL,
	UNIQUE (name)
);

CREATE TABLE domain (
	id uuid PRIMARY KEY,
	created_at timestamp WITHOUT TIME ZONE NOT NULL,
	updated_at timestamp WITHOUT TIME ZONE NOT NULL,
	domain text NOT NULL,
	app_id uuid REFERENCES app(id) NOT NULL
);
