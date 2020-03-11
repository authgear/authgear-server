-- Put upgrade SQL here
CREATE TABLE cloud_code (
	id uuid PRIMARY KEY,
	created_at timestamp WITHOUT TIME ZONE NOT NULL,
	version text NOT NULL,
	path text NOT NULL,
	target_path text NOT NULL,
	config jsonb NOT NULL,
	app_id uuid REFERENCES app(id) NOT NULL
);
