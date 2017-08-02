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

type revision_b55e91bc9391 struct {
}

func (r *revision_b55e91bc9391) Version() string {
	return "b55e91bc9391"
}

func (r *revision_b55e91bc9391) Up(tx *sqlx.Tx) error {
	stmt := `
INSERT INTO _record_field_access
  (record_type, record_field, user_role, writable, readable, comparable, discoverable)
VALUES
  ('user', 'username', '_any_user', 'FALSE', 'FALSE', 'FALSE', 'TRUE');

INSERT INTO _record_field_access
  (record_type, record_field, user_role, writable, readable, comparable, discoverable)
VALUES
  ('user', 'username', '_owner', 'TRUE', 'TRUE', 'TRUE', 'TRUE');

INSERT INTO _record_field_access
  (record_type, record_field, user_role, writable, readable, comparable, discoverable)
VALUES
  ('user', 'email', '_any_user', 'FALSE', 'FALSE', 'FALSE', 'TRUE');

INSERT INTO _record_field_access
  (record_type, record_field, user_role, writable, readable, comparable, discoverable)
VALUES
  ('user', 'email', '_owner', 'TRUE', 'TRUE', 'TRUE', 'TRUE');
  `

	_, err := tx.Exec(stmt)
	return err
}

func (r *revision_b55e91bc9391) Down(tx *sqlx.Tx) error {
	// no op
	return nil
}
