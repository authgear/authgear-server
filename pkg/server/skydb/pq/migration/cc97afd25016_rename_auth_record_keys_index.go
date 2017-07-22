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

type revision_cc97afd25016 struct {
}

func (r *revision_cc97afd25016) Version() string {
	return "cc97afd25016"
}

func (r *revision_cc97afd25016) Up(tx *sqlx.Tx) error {
	stmt := `
ALTER INDEX user_username_key RENAME TO auth_record_keys_user_username_key;
ALTER INDEX user_email_key RENAME TO auth_record_keys_user_email_key;
  `

	_, err := tx.Exec(stmt)
	return err
}

func (r *revision_cc97afd25016) Down(tx *sqlx.Tx) error {
	var err error

	// drop if exist
	tx.Exec(`ALTER TABLE "user" DROP CONSTRAINT auth_record_keys_user_username_key;`)
	tx.Exec(`ALTER TABLE "user" DROP CONSTRAINT auth_record_keys_user_email_key;`)

	stmt := `
ALTER TABLE "user" ADD UNIQUE (username);
ALTER TABLE "user" ADD UNIQUE (email);
  `

	_, err = tx.Exec(stmt)
	return err
}
