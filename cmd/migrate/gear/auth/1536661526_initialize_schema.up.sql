BEGIN;

CREATE TABLE auth_principal (
  id text PRIMARY KEY, -- anonymous|abc, password|abc, custom_token|abc, oauth:facebook|abc, custom:code|abc
  provider text NOT NULL, -- anonymous, password, custom_token, oauth:facebook, oauth:google, custom:code
  user_id text REFERENCES _user(id),
  CONSTRAINT _auth_principal_user_id_provider_key UNIQUE (user_id, provider)
);

CREATE TABLE auth_provider_password (
  principal_id text PRIMARY KEY REFERENCES auth_principal(id),
  auth_data jsonb,
  password text NOT NULL
);

CREATE TABLE auth_provider_oauth (
  principal_id text PRIMARY KEY REFERENCES auth_principal(id),
  oauth_provider text NOT NULL, -- facebook, google
  token_response jsonb,
  profile jsonb,
  _created_at timestamp without time zone NOT NULL,
  _updated_at timestamp without time zone NOT NULL
);

CREATE TABLE auth_password_history (
  id text PRIMARY KEY,
  user_id text NOT NULL,
  password text NOT NULL,
  logged_at timestamp without time zone NOT NULL
);

END;
