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

type revision_30d0a626888 struct {
}

func (r *revision_30d0a626888) Version() string { return "30d0a626888" }

func (r *revision_30d0a626888) Up(tx *sqlx.Tx) error {
	stmts := []string{
		`ALTER TABLE _user ADD COLUMN username VARCHAR(255);`,
		`ALTER TABLE _user ADD UNIQUE (username);`,
		`ALTER TABLE _user ADD CONSTRAINT _user_email_key UNIQUE (email);`,
	}
	for _, stmt := range stmts {
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (r *revision_30d0a626888) Down(tx *sqlx.Tx) error {
	stmts := []string{
		`ALTER TABLE _user DROP COLUMN username;`,
		`ALTER TABLE _user DROP CONSTRAINT _user_email_key;`,
	}
	for _, stmt := range stmts {
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
