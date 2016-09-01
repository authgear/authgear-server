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

type revision_db76e79e987 struct {
}

func (r *revision_db76e79e987) Version() string { return "db76e79e987" }

func (r *revision_db76e79e987) Up(tx *sqlx.Tx) error {
	stmts := []string{
		`ALTER TABLE _user ADD COLUMN last_login_at timestamp without time zone;`,
		`ALTER TABLE _user ADD COLUMN last_seen_at timestamp without time zone;`,
	}
	for _, stmt := range stmts {
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (r *revision_db76e79e987) Down(tx *sqlx.Tx) error {
	stmts := []string{
		`ALTER TABLE _user DROP COLUMN last_login_at;`,
		`ALTER TABLE _user DROP COLUMN last_seen_at;`,
	}
	for _, stmt := range stmts {
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
