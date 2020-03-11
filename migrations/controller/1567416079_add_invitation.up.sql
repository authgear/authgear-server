CREATE TABLE invitation (
	id uuid PRIMARY KEY,
    created_by text REFERENCES _core_user(id),
	created_at timestamp WITHOUT TIME ZONE NOT NULL,
    email text NOT NULL,
	app_id uuid REFERENCES app(id) NOT NULL,
    code text NOT NULL,
	consumed boolean NOT NULL DEFAULT FALSE,
	expiry timestamp WITHOUT TIME ZONE NOT NULL
);
