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

type revision_6250fd676f58 struct {
}

func (r *revision_6250fd676f58) Version() string {
	return "6250fd676f58"
}

func (r *revision_6250fd676f58) Up(tx *sqlx.Tx) error {
	stmt := `
CREATE TABLE _sso_custom_token (
  user_id text NOT NULL PRIMARY KEY,
  principal_id text NOT NULL,
  _created_at timestamp without time zone NOT NULL,
  UNIQUE (principal_id)
);`

	_, err := tx.Exec(stmt)
	return err
}

func (r *revision_6250fd676f58) Down(tx *sqlx.Tx) error {
	stmt := `DROP TABLE _sso_custom_token;`

	_, err := tx.Exec(stmt)
	return err
}
