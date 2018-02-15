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

type revision_b3163d49bd6d struct {
}

func (r *revision_b3163d49bd6d) Version() string {
	return "b3163d49bd6d"
}

func (r *revision_b3163d49bd6d) Up(tx *sqlx.Tx) error {
	stmt := `
	ALTER TABLE _auth ADD COLUMN disabled boolean NOT NULL DEFAULT FALSE;
	ALTER TABLE _auth ADD COLUMN disabled_message text;
	ALTER TABLE _auth ADD COLUMN disabled_expiry timestamp without time zone;
	`
	_, err := tx.Exec(stmt)
	return err
}

func (r *revision_b3163d49bd6d) Down(tx *sqlx.Tx) error {
	stmt := `
	ALTER TABLE _auth DROP COLUMN disabled;
	ALTER TABLE _auth DROP COLUMN disabled_message;
	ALTER TABLE _auth DROP COLUMN disabled_expiry;
	`
	_, err := tx.Exec(stmt)
	return err
}
