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

type revision_83f549ff247b struct {
}

func (r *revision_83f549ff247b) Version() string {
	return "83f549ff247b"
}

func (r *revision_83f549ff247b) Up(tx *sqlx.Tx) error {
	stmt := `
		ALTER TABLE "user"
		  ADD COLUMN last_login_at timestamp without time zone;

		UPDATE "user" AS u
		SET last_login_at = a.last_login_at
		FROM _auth AS a
		WHERE a.id = u._id;

		CREATE OR REPLACE VIEW _user AS
		  SELECT
		    a.id,
		    a.password,
		    u.username,
		    u.email,
		    a.provider_info AS auth,
		    a.token_valid_since,
		    u.last_login_at,
		    a.last_seen_at
		   FROM _auth as a
		     JOIN "user" AS u ON u._id = a.id;

		ALTER TABLE _auth
		  DROP COLUMN last_login_at;
	`

	_, err := tx.Exec(stmt)
	return err
}

func (r *revision_83f549ff247b) Down(tx *sqlx.Tx) error {
	stmt := `
		ALTER TABLE _auth
		  ADD COLUMN last_login_at timestamp without time zone;

		UPDATE _auth AS a
		SET last_login_at = u.last_login_at
		FROM "user" AS u
		WHERE a.id = u._id;

		CREATE OR REPLACE VIEW _user AS
		  SELECT
		    a.id,
		    a.password,
		    u.username,
		    u.email,
		    a.provider_info AS auth,
		    a.token_valid_since,
		    a.last_login_at,
		    a.last_seen_at
		   FROM _auth as a
		     JOIN "user" AS u ON u._id = a.id;

		ALTER TABLE "user"
		  DROP COLUMN last_login_at;
	`

	_, err := tx.Exec(stmt)
	return err
}
