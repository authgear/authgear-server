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

type revision_d2cb54c648 struct {
}

func (r *revision_d2cb54c648) Version() string { return "d2cb54c648" }

func (r *revision_d2cb54c648) Up(tx *sqlx.Tx) error {
	tables, err := getAllRecordTables(tx)
	if err != nil {
		return err

	}
	stmts := []string{
		`ALTER TABLE %s ADD COLUMN _created_at TIMESTAMP WITHOUT TIME ZONE;`,
		`ALTER TABLE %s ADD COLUMN _updated_at TIMESTAMP WITHOUT TIME ZONE;`,
		`UPDATE %s SET _created_at = now() at time zone 'utc', _updated_at = now() at time zone 'utc';`,
		`ALTER TABLE %s ALTER COLUMN _created_at SET NOT NULL;`,
		`ALTER TABLE %s ALTER COLUMN _updated_at SET NOT NULL;`,
		`ALTER TABLE %s ADD COLUMN _created_by TEXT;`,
		`ALTER TABLE %s ADD COLUMN _updated_by TEXT;`,
	}
	for _, name := range tables {
		for _, stmt := range stmts {
			_, err = tx.Exec(fmt.Sprintf(stmt, name))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *revision_d2cb54c648) Down(tx *sqlx.Tx) error {
	tables, err := getAllRecordTables(tx)
	if err != nil {
		return err

	}
	stmts := []string{
		`ALTER TABLE %s DROP COLUMN _created_at;`,
		`ALTER TABLE %s DROP COLUMN _created_by;`,
		`ALTER TABLE %s DROP COLUMN _updated_at;`,
		`ALTER TABLE %s DROP COLUMN _updated_by;`,
	}
	for _, name := range tables {
		for _, stmt := range stmts {
			_, err = tx.Exec(fmt.Sprintf(stmt, name))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
