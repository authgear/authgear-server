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

type revision_81beb4d8658c struct {
}

func (r *revision_81beb4d8658c) Version() string {
	return "81beb4d8658c"
}

func (r *revision_81beb4d8658c) Up(tx *sqlx.Tx) error {
	stmt := `
    DELETE FROM _record_field_access
    WHERE record_type = 'user'
      AND record_field = 'last_login_at'
      AND user_role = '_owner';

    DELETE FROM _record_field_access
    WHERE record_type = 'user'
      AND record_field = 'last_login_at'
      AND user_role = '_public';

    INSERT INTO _record_field_access
      (record_type, record_field, user_role, writable, readable, comparable, discoverable)
    VALUES
      ('user', 'last_login_at', '_owner', FALSE, TRUE, FALSE, FALSE),
      ('user', 'last_login_at', '_public', FALSE, FALSE, FALSE, FALSE);
  `

	_, err := tx.Exec(stmt)
	return err
}

func (r *revision_81beb4d8658c) Down(tx *sqlx.Tx) error {
	stmt := `
    DELETE FROM _record_field_access
    WHERE record_type = 'user'
      AND record_field = 'last_login_at'
      AND user_role = '_owner';

    DELETE FROM _record_field_access
    WHERE record_type = 'user'
      AND record_field = 'last_login_at'
      AND user_role = '_public';
  `

	_, err := tx.Exec(stmt)
	return err
}
