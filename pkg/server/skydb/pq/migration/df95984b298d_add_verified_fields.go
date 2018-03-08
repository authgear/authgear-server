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

type revision_df95984b298d struct {
}

func (r *revision_df95984b298d) Version() string {
	return "df95984b298d"
}

func (r *revision_df95984b298d) Up(tx *sqlx.Tx) error {
	stmt := `
	ALTER TABLE _auth ADD COLUMN verified boolean NOT NULL DEFAULT FALSE;
	`
	_, err := tx.Exec(stmt)
	return err
}

func (r *revision_df95984b298d) Down(tx *sqlx.Tx) error {
	stmt := `
	ALTER TABLE _auth DROP COLUMN verified;
	ALTER TABLE _auth DROP COLUMN disabled_message;
	ALTER TABLE _auth DROP COLUMN disabled_expiry;
	`
	_, err := tx.Exec(stmt)
	return err
}
