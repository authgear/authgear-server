package migration

import (
	"github.com/jmoiron/sqlx"
)

type revision_bc5768da91b struct {
}

func (r *revision_bc5768da91b) Version() string { return "bc5768da91b" }

func (r *revision_bc5768da91b) Up(tx *sqlx.Tx) error {
	stmt := `
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
		`

	_, err := tx.Exec(stmt)
	return err
}

func (r *revision_bc5768da91b) Down(tx *sqlx.Tx) error {
	stmt := `
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
		`

	_, err := tx.Exec(stmt)
	return err
}
