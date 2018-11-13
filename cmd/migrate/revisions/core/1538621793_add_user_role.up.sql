BEGIN;

CREATE TABLE _core_role (
  id text PRIMARY KEY,
  by_default boolean NOT NULL DEFAULT false,
  is_admin boolean NOT NULL DEFAULT false
);

CREATE TABLE _core_user_role (
  user_id text REFERENCES _core_user(id),
  role_id text REFERENCES _core_role(id),
  CONSTRAINT _auth_role_pkey PRIMARY KEY (user_id, role_id)
);

END;
