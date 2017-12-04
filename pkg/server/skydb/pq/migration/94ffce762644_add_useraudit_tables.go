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

type revision_94ffce762644 struct {
}

func (r *revision_94ffce762644) Version() string {
	return "94ffce762644"
}

func (r *revision_94ffce762644) Up(tx *sqlx.Tx) error {
	stmt := `
	CREATE TABLE _password_history (
		id TEXT PRIMARY KEY,
		auth_id TEXT NOT NULL,
		password TEXT NOT NULL,
		logged_at TIMESTAMP WITHOUT TIME ZONE NOT NULL
	);
	CREATE INDEX ON _password_history (auth_id, logged_at DESC);
	`
	_, err := tx.Exec(stmt)
	return err
}

func (r *revision_94ffce762644) Down(tx *sqlx.Tx) error {
	stmt := `
	DROP TABLE _password_history;
	`
	_, err := tx.Exec(stmt)
	return err
}
