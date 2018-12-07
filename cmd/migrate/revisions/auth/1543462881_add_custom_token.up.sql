CREATE TABLE _auth_provider_custom_token (
  principal_id text PRIMARY KEY REFERENCES _auth_principal(id),
  token_principal_id text NOT NULL,
  UNIQUE (principal_id),
  UNIQUE (token_principal_id)
);
