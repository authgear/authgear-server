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

import "github.com/jmoiron/sqlx"

type revision_7469be11899e struct {
}

func (r *revision_7469be11899e) Version() string {
	return "7469be11899e"
}

func (r *revision_7469be11899e) Up(tx *sqlx.Tx) error {
	stmt := `
	CREATE TABLE _verify_code (
		id TEXT PRIMARY KEY,
		auth_id TEXT NOT NULL,
		record_key TEXT NOT NULL,
		record_value TEXT NOT NULL,
		code TEXT NOT NULL,
		consumed BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL
	);
	CREATE INDEX ON _verify_code (auth_id, code, consumed);
	`
	_, err := tx.Exec(stmt)
	return err
}

func (r *revision_7469be11899e) Down(tx *sqlx.Tx) error {
	stmt := `
	DROP TABLE _verify_code;
	`
	_, err := tx.Exec(stmt)
	return err
}
