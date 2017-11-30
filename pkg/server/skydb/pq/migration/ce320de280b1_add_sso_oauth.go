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

type revision_ce320de280b1 struct {
}

func (r *revision_ce320de280b1) Version() string {
	return "ce320de280b1"
}

func (r *revision_ce320de280b1) Up(tx *sqlx.Tx) error {
	stmt := `CREATE TABLE _sso_oauth (
    user_id text NOT NULL,
    provider text NOT NULL,
    principal_id text NOT NULL,
    token_response jsonb,
    profile jsonb,
    _created_at timestamp without time zone NOT NULL,
    _updated_at timestamp without time zone NOT NULL,
    PRIMARY KEY (provider, principal_id),
  	UNIQUE (user_id, provider)
  );`

	_, err := tx.Exec(stmt)
	return err
}

func (r *revision_ce320de280b1) Down(tx *sqlx.Tx) error {
	stmt := `DROP TABLE _sso_oauth;`

	_, err := tx.Exec(stmt)
	return err
}
