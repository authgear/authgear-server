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

type revision_41af1c8d394 struct {
}

func (r *revision_41af1c8d394) Version() string { return "41af1c8d394" }

func (r *revision_41af1c8d394) Up(tx *sqlx.Tx) error {

	stmts := []string{
		`
		CREATE TABLE _role (
		    id TEXT NOT NULL,
		    PRIMARY KEY (id)
		);
		`,
		`
		CREATE TABLE _user_role (
		    user_id TEXT NOT NULL,
		    role_id TEXT NOT NULL,
		    UNIQUE (user_id, role_id),
		    FOREIGN KEY(user_id) REFERENCES _user (id),
		    FOREIGN KEY(role_id) REFERENCES _role (id)
		);
		`,
	}
	for _, stmt := range stmts {
		_, err := tx.Exec(stmt)
		return err
	}
	return nil
}

func (r *revision_41af1c8d394) Down(tx *sqlx.Tx) error {
	stmts := []string{
		`DROP TABLE _user_role;`,
		`DROP TABLE _role;`,
	}
	for _, stmt := range stmts {
		_, err := tx.Exec(stmt)
		return err
	}
	return nil
}
