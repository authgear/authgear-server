CREATE TABLE _auth_user_profile (
  user_id text REFERENCES _core_user(id),
  created_at timestamp without time zone NOT NULL,
  created_by text,
  updated_at timestamp without time zone NOT NULL,
  updated_by text,
  data jsonb,
  PRIMARY KEY(user_id),
  UNIQUE (user_id)
);
