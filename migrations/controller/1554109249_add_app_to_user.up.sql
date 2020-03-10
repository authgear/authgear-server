CREATE TABLE app_user (
    user_id text REFERENCES _core_user(id),
    app_id uuid REFERENCES app(id),
    created_at timestamp WITHOUT TIME ZONE NOT NULL,
    CONSTRAINT app_user_pkey PRIMARY KEY (user_id, app_id)
);
