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
