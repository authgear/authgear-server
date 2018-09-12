BEGIN;

CREATE TABLE _user (
  id text PRIMARY KEY,
  token_valid_since timestamp without time zone,
  last_seen_at timestamp without time zone,
  disabled boolean NOT NULL DEFAULT false,
  disabled_message text,
  disabled_expiry timestamp without time zone
);

END;
