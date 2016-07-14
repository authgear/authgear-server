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
	"github.com/jmoiron/sqlx"
)

type fullMigration struct {
}

func (r *fullMigration) Version() string { return "88a550bf579" }

func (r *fullMigration) Up(tx *sqlx.Tx) error {
	const stmt = `
CREATE EXTENSION IF NOT EXISTS citext;
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

CREATE TABLE _user (
	id text PRIMARY KEY,
	username citext,
	email citext,
	password text,
	auth jsonb,
	token_valid_since timestamp without time zone,
	UNIQUE (username),
	UNIQUE (email)
);

CREATE TABLE _role (
	id text PRIMARY KEY,
	by_default boolean DEFAULT FALSE,
	is_admin boolean DEFAULT FALSE
);

CREATE TABLE _user_role (
	user_id text REFERENCES _user (id) NOT NULL,
	role_id text REFERENCES _role (id) NOT NULL,
	PRIMARY KEY (user_id, role_id)
);

CREATE TABLE _asset (
	id text PRIMARY KEY,
	content_type text NOT NULL,
	size bigint NOT NULL
);
CREATE TABLE _device (
	id text PRIMARY KEY,
	user_id text REFERENCES _user (id),
	type text NOT NULL,
	token text,
	last_registered_at timestamp without time zone NOT NULL,
	UNIQUE (user_id, type, token)
);
CREATE INDEX ON _device (token, last_registered_at);
CREATE TABLE _subscription (
	id text NOT NULL,
	user_id text NOT NULL,
	device_id text REFERENCES _device (id) ON DELETE CASCADE NOT NULL,
	type text NOT NULL,
	notification_info jsonb,
	query jsonb,
	PRIMARY KEY(user_id, device_id, id)
);
CREATE TABLE _friend (
	left_id text NOT NULL,
	right_id text REFERENCES _user (id) NOT NULL,
	PRIMARY KEY(left_id, right_id)
);
CREATE TABLE _follow (
	left_id text NOT NULL,
	right_id text REFERENCES _user (id) NOT NULL,
	PRIMARY KEY(left_id, right_id)
);
CREATE TABLE _record_creation (
    record_type text NOT NULL,
    role_id text,
    UNIQUE (record_type, role_id),
    FOREIGN KEY (role_id) REFERENCES _role(id)
);
CREATE INDEX _record_creation_unique_record_type ON _record_creation (record_type);
`
	_, err := tx.Exec(stmt)
	return err
}

func (r *fullMigration) Down(tx *sqlx.Tx) error {
	panic("cannot downgrade from a base revision")
}
