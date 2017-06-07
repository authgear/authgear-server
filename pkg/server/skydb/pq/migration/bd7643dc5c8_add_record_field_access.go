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

type revision_bd7643dc5c8 struct {
}

func (r *revision_bd7643dc5c8) Version() string { return "bd7643dc5c8" }

func (r *revision_bd7643dc5c8) Up(tx *sqlx.Tx) error {
	stmts := []string{
		`CREATE TABLE _record_field_access (
			record_type text NOT NULL,
			record_field text NOT NULL,
			user_role text NOT NULL,
			writable boolean NOT NULL,
			readable boolean NOT NULL,
			comparable boolean NOT NULL,
			discoverable boolean NOT NULL,
			PRIMARY KEY (record_type, record_field, user_role)
		);`,
	}
	for _, stmt := range stmts {
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (r *revision_bd7643dc5c8) Down(tx *sqlx.Tx) error {
	stmts := []string{
		`DROP TABLE _record_field_access;`,
	}
	for _, stmt := range stmts {
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
