package migration

import (
	"bytes"
	"text/template"

	"github.com/jmoiron/sqlx"
)

var createAppSchemaStmtTmpl *template.Template

func init() {
	createAppSchemaStmtTmpl = template.Must(template.New("createAppSchemaStmtTmpl").Parse(createAppSchemaStmtTmplText))
}

func InitSchema(tx *sqlx.Tx, schema string) error {
	stmt, err := templateExecString(createAppSchemaStmtTmpl, struct {
		Schema     string
		VersionNum string
	}{schema, DbVersionNum})

	if err != nil {
		return err
	}

	_, err = tx.Exec(stmt)
	return err
}

func templateExecString(t *template.Template, i interface{}) (string, error) {
	var buf bytes.Buffer
	if err := t.Execute(&buf, i); err != nil {
		return "", err
	}

	return buf.String(), nil
}

const DbVersionNum = "551bc42a839"
const createAppSchemaStmtTmplText = `
CREATE SCHEMA IF NOT EXISTS {{.Schema}};
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
		INSERT INTO pending_notification (op, appname, recordtype, record)
			VALUES (TG_OP, TG_TABLE_SCHEMA, TG_TABLE_NAME, row_to_json(affected_record)::jsonb)
			RETURNING id INTO inserted_id;
		PERFORM pg_notify('record_change', inserted_id::TEXT);
		RETURN affected_record;
	END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS {{.Schema}}._version (
	version_num character varying(32) NOT NULL
);
INSERT INTO {{.Schema}}._version (version_num) VALUES('{{.VersionNum}}');

CREATE TABLE {{.Schema}}._user (
	id text PRIMARY KEY,
	username text,
	email text,
	password text,
	auth jsonb,
	UNIQUE (username),
	UNIQUE (email)
);

CREATE TABLE {{.Schema}}._role (
	id text PRIMARY KEY,
	by_default boolean DEFAULT FALSE,
	is_admin boolean DEFAULT FALSE
);

CREATE TABLE {{.Schema}}._user_role (
	user_id text REFERENCES {{.Schema}}._user (id) NOT NULL,
	role_id text REFERENCES {{.Schema}}._role (id) NOT NULL,
	PRIMARY KEY (user_id, role_id)
);

CREATE TABLE {{.Schema}}._asset (
	id text PRIMARY KEY,
	content_type text NOT NULL,
	size bigint NOT NULL
);
CREATE TABLE {{.Schema}}._device (
	id text PRIMARY KEY,
	user_id text REFERENCES {{.Schema}}._user (id),
	type text NOT NULL,
	token text,
	last_registered_at timestamp without time zone NOT NULL,
	UNIQUE (user_id, type, token)
);
CREATE INDEX ON {{.Schema}}._device (token, last_registered_at);
CREATE TABLE {{.Schema}}._subscription (
	id text NOT NULL,
	user_id text NOT NULL,
	device_id text REFERENCES {{.Schema}}._device (id) ON DELETE CASCADE NOT NULL,
	type text NOT NULL,
	notification_info jsonb,
	query jsonb,
	PRIMARY KEY(user_id, device_id, id)
);
CREATE TABLE {{.Schema}}._friend (
	left_id text NOT NULL,
	right_id text REFERENCES {{.Schema}}._user (id) NOT NULL,
	PRIMARY KEY(left_id, right_id)
);
CREATE TABLE {{.Schema}}._follow (
	left_id text NOT NULL,
	right_id text REFERENCES {{.Schema}}._user (id) NOT NULL,
	PRIMARY KEY(left_id, right_id)
);
`
