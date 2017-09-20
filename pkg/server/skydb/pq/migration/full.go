// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package migration

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type fullMigration struct {
}

func (r *fullMigration) Version() string { return "b55e91bc9391" }

func (r *fullMigration) createTable(tx *sqlx.Tx) error {
	const stmt = `
CREATE EXTENSION IF NOT EXISTS postgis WITH SCHEMA public;
CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA public;
CREATE TABLE IF NOT EXISTS public.pending_notification (
	id SERIAL NOT NULL PRIMARY KEY,
	op text NOT NULL,
	appname text NOT NULL,
	recordtype text NOT NULL,
	record jsonb NOT NULL
);
CREATE OR REPLACE FUNCTION public.notify_record_change() RETURNS TRIGGER AS $$
	DECLARE
		affected_record RECORD;
		inserted_id integer;
	BEGIN
		IF (TG_OP = 'DELETE') THEN
			affected_record := OLD;
		ELSE
			affected_record := NEW;
		END IF;
		INSERT INTO public.pending_notification (op, appname, recordtype, record)
			VALUES (TG_OP, TG_TABLE_SCHEMA, TG_TABLE_NAME, row_to_json(affected_record)::jsonb)
			RETURNING id INTO inserted_id;
		PERFORM pg_notify('record_change', inserted_id::TEXT);
		RETURN affected_record;
	END;
$$ LANGUAGE plpgsql;

CREATE TABLE _auth (
	id text PRIMARY KEY,
	password text,
	provider_info jsonb,
	token_valid_since timestamp without time zone,
	last_seen_at timestamp without time zone
);

CREATE TABLE _role (
	id text PRIMARY KEY,
	by_default boolean DEFAULT FALSE,
	is_admin boolean DEFAULT FALSE
);

CREATE TABLE _auth_role (
	auth_id text REFERENCES _auth (id) NOT NULL,
	role_id text REFERENCES _role (id) NOT NULL,
	PRIMARY KEY (auth_id, role_id)
);

CREATE TABLE _asset (
	id text PRIMARY KEY,
	content_type text NOT NULL,
	size bigint NOT NULL
);
CREATE TABLE _device (
	id text PRIMARY KEY,
	auth_id text REFERENCES _auth (id),
	type text NOT NULL,
	token text,
	topic text,
	last_registered_at timestamp without time zone NOT NULL,
	UNIQUE (auth_id, type, token)
);
CREATE INDEX ON _device (token, last_registered_at);
CREATE TABLE _subscription (
	id text NOT NULL,
	auth_id text NOT NULL,
	device_id text REFERENCES _device (id) ON DELETE CASCADE NOT NULL,
	type text NOT NULL,
	notification_info jsonb,
	query jsonb,
	PRIMARY KEY(auth_id, device_id, id)
);
CREATE TABLE _friend (
	left_id text NOT NULL,
	right_id text REFERENCES _auth (id) NOT NULL,
	PRIMARY KEY(left_id, right_id)
);
CREATE TABLE _follow (
	left_id text NOT NULL,
	right_id text REFERENCES _auth (id) NOT NULL,
	PRIMARY KEY(left_id, right_id)
);
CREATE TABLE _record_creation (
    record_type text NOT NULL,
    role_id text,
    UNIQUE (record_type, role_id),
    FOREIGN KEY (role_id) REFERENCES _role(id)
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
CREATE TABLE "user" (
    _id text,
    _database_id text,
    _owner_id text,
    _access jsonb,
    _created_at timestamp without time zone NOT NULL,
    _created_by text,
    _updated_at timestamp without time zone NOT NULL,
    _updated_by text,
    username citext,
    email citext,
    last_login_at timestamp without time zone,
    PRIMARY KEY(_id, _database_id, _owner_id),
    UNIQUE (_id)
);
ALTER TABLE "user" ADD CONSTRAINT auth_record_keys_user_username_key UNIQUE (username);
ALTER TABLE "user" ADD CONSTRAINT auth_record_keys_user_email_key UNIQUE (email);
CREATE VIEW _user AS
    SELECT
        a.id,
        a.password,
        u.username,
        u.email,
        a.provider_info AS auth,
        a.token_valid_since,
        u.last_login_at,
        a.last_seen_at
    FROM _auth AS a
    JOIN "user" AS u ON u._id = a.id;

INSERT INTO _record_field_access
  (record_type, record_field, user_role, writable, readable, comparable, discoverable)
VALUES
  ('user', 'username', '_any_user', 'FALSE', 'TRUE', 'FALSE', 'TRUE');

INSERT INTO _record_field_access
  (record_type, record_field, user_role, writable, readable, comparable, discoverable)
VALUES
  ('user', 'username', '_owner', 'TRUE', 'TRUE', 'TRUE', 'TRUE');

INSERT INTO _record_field_access
  (record_type, record_field, user_role, writable, readable, comparable, discoverable)
VALUES
  ('user', 'email', '_any_user', 'FALSE', 'TRUE', 'FALSE', 'TRUE');

INSERT INTO _record_field_access
  (record_type, record_field, user_role, writable, readable, comparable, discoverable)
VALUES
  ('user', 'email', '_owner', 'TRUE', 'TRUE', 'TRUE', 'TRUE');
`
	_, err := tx.Exec(stmt)
	return err
}

func (r *fullMigration) insertSeedData(tx *sqlx.Tx) error {
	stmts := []string{
		fmt.Sprintf(
			`INSERT INTO _role (id, is_admin) VALUES ('%s', TRUE)`,
			adminRoleDefaultName,
		),
	}

	for _, stmt := range stmts {
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}

func (r *fullMigration) Up(tx *sqlx.Tx) error {
	var err error
	if err = r.createTable(tx); err != nil {
		return err
	}

	return r.insertSeedData(tx)
}

func (r *fullMigration) Down(tx *sqlx.Tx) error {
	panic("cannot downgrade from a base revision")
}
