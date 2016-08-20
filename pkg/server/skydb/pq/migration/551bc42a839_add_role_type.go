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

type revision_551bc42a839 struct {
}

func (r *revision_551bc42a839) Version() string { return "551bc42a839" }

func (r *revision_551bc42a839) Up(tx *sqlx.Tx) error {
	stmts := []string{
		`ALTER TABLE _role ADD COLUMN by_default boolean DEFAULT FALSE;`,
		`ALTER TABLE _role ADD COLUMN is_admin boolean DEFAULT FALSE;`,
	}
	for _, stmt := range stmts {
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (r *revision_551bc42a839) Down(tx *sqlx.Tx) error {
	stmts := []string{
		`ALTER TABLE _role DROP COLUMN is_admin;`,
		`ALTER TABLE _role DROP COLUMN by_default;`,
	}
	for _, stmt := range stmts {
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
