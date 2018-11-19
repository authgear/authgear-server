CREATE TABLE _record_asset (
  id text PRIMARY KEY,
  content_type text NOT NULL,
  size bigint NOT NULL
);

CREATE TABLE _record_creation (
  record_type text NOT NULL,
  role_id text,
  UNIQUE (record_type, role_id),
  FOREIGN KEY (role_id) REFERENCES _core_role(id)
);
CREATE INDEX _record_creation_unique_record_type ON _record_creation (record_type);

CREATE TABLE _record_default_access (
  record_type text NOT NULL,
  default_access jsonb,
  UNIQUE (record_type)
);
CREATE INDEX _record_default_access_unique_record_type ON _record_default_access (record_type);

CREATE TABLE _record_field_access (
  record_type text NOT NULL,
  record_field text NOT NULL,
  user_role text NOT NULL,
  writable boolean NOT NULL,
  readable boolean NOT NULL,
  comparable boolean NOT NULL,
  discoverable boolean NOT NULL,
  PRIMARY KEY (record_type, record_field, user_role)
);
