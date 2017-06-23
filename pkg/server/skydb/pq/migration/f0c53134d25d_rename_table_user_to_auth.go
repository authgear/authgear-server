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
	"fmt"

  "github.com/jmoiron/sqlx"
)

type ReservedColumnExistedError struct {
	table  string
	column string
}

func (e *ReservedColumnExistedError) Error() string {
	// TODO: put a link of guide here for those who failed to migrate
	return fmt.Sprintf(`Reserved column %s.%s has already been existed`, e.table, e.column)
}

type revision_f0c53134d25d struct {
}

func (r *revision_f0c53134d25d) Version() string {
	return "f0c53134d25d"
}

func (r *revision_f0c53134d25d) IsTableExisted(tx *sqlx.Tx, table string) (bool, error) {
	var exists bool
	err := tx.QueryRowx(`
SELECT EXISTS (
	SELECT 1
	FROM information_schema.tables
	WHERE
		table_schema = current_schema() AND
		table_name = $1
);
		`, table).Scan(&exists)
	return exists, err
}

func (r *revision_f0c53134d25d) IsColumnExisted(tx *sqlx.Tx, table string, column string) (bool, error) {
	var exists bool
	err := tx.QueryRowx(`
SELECT EXISTS (
	SELECT 1
	FROM information_schema.columns
	WHERE
		table_schema = current_schema() AND
		table_name = $1 AND
		column_name = $2
);
		`, table, column).Scan(&exists)
	return exists, err
}

func (r *revision_f0c53134d25d) EnsureReservedColumnsNotExisted(tx *sqlx.Tx) error {
	table := `user`
	columns := [2]string{`username`, `email`}

	for _, column := range columns {
		isExisted, err := r.IsColumnExisted(tx, table, column)
		if err != nil {
			return err
		}

		if isExisted {
			return &ReservedColumnExistedError{table: table, column: column}
		}
	}

	return nil
}

func (r *revision_f0c53134d25d) Up(tx *sqlx.Tx) error {
	migrateSkygearUserStmt := `
-- Migrate _user to _auth
ALTER TABLE _user_role DROP CONSTRAINT _user_role_user_id_fkey;
ALTER TABLE _device DROP CONSTRAINT _device_user_id_fkey;
ALTER TABLE _friend DROP CONSTRAINT _friend_right_id_fkey;
ALTER TABLE _follow DROP CONSTRAINT _follow_right_id_fkey;

ALTER TABLE _user_role RENAME TO _auth_role;

ALTER TABLE _user RENAME TO _auth;

ALTER TABLE _auth_role RENAME user_id TO auth_id;
ALTER TABLE _device RENAME user_id TO auth_id;
ALTER TABLE _subscription RENAME user_id TO auth_id;

ALTER TABLE _auth_role
	ADD CONSTRAINT _auth_role_auth_id_fkey FOREIGN KEY (auth_id) REFERENCES _auth (id);
ALTER TABLE _device
	ADD CONSTRAINT _device_auth_id_fkey FOREIGN KEY (auth_id) REFERENCES _auth (id);
ALTER TABLE _friend
	ADD CONSTRAINT _friend_right_id_fkey FOREIGN KEY (right_id) REFERENCES _auth (id);
ALTER TABLE _follow
	ADD CONSTRAINT _follow_right_id_fkey FOREIGN KEY (right_id) REFERENCES _auth (id);

CREATE VIEW _user AS
SELECT * FROM _auth;

ALTER TABLE _auth RENAME auth to provider_info;
	 `
	_, err := tx.Exec(migrateSkygearUserStmt)
	if err != nil {
		return err
	}

	if err = r.EnsureReservedColumnsNotExisted(tx); err != nil {
		return err
	}

	var userTableExists bool
	userTableExists, err = r.IsTableExisted(tx, `user`)
	if err != nil {
		return err
	}

	var migrateUserStmt string
	if userTableExists {
		migrateUserStmt = `
-- Migrate username and email to user table
ALTER TABLE user ADD COLUMN username citext UNIQUE;
ALTER TABLE user ADD COLUMN email citext UNIQUE;

UPDATE user
SET
	username = a.username,
	email = a.email
FROM _auth as a;

ALTER TABLE _auth DROP COLUMN username citext UNIQUE;
ALTER TABLE _auth DROP COLUMN email citext UNIQUE;
		`
	} else {
		migrateUserStmt = `
-- Create user table if not existed
CREATE TABLE user (
    _id text,
    _database_id text,
    _owner_id text,
    _access jsonb,
    _created_at timestamp without time zone NOT NULL,
    _created_by text,
    _updated_at timestamp without time zone NOT NULL,
    _updated_by text,
    username citext,
    email citext,
    PRIMARY KEY(_id, _database_id, _owner_id),
    UNIQUE (_id),
    UNIQUE (username),
    UNIQUE (email)
);
		`
	}

	_, err = tx.Exec(migrateUserStmt)
	if err != nil {
		return err
	}

	return nil
}

func (r *revision_f0c53134d25d) Down(tx *sqlx.Tx) error {
	migrateUserStmt := `
-- Migrate username and email to user table (backward)
ALTER TABLE _auth ADD COLUMN username citext UNIQUE;
ALTER TABLE _auth ADD COLUMN email citext UNIQUE;

UPDATE _auth
SET
	username = u.username,
	email = u.email
FROM user as u;

ALTER TABLE user DROP COLUMN username citext UNIQUE;
ALTER TABLE user DROP COLUMN email citext UNIQUE;
		`
	_, err := tx.Exec(migrateUserStmt)
	if err != nil {
		return err
	}

	migrateSkygearUserStmt := `
-- Migrate _user to _auth (backward)
ALTER TABLE _auth RENAME provider_info to auth;

DROP VIEW _user;

ALTER TABLE _auth_role DROP CONSTRAINT _auth_role_auth_id_fkey;
ALTER TABLE _device DROP CONSTRAINT _device_auth_id_fkey;
ALTER TABLE _friend DROP CONSTRAINT _friend_right_id_fkey;
ALTER TABLE _follow DROP CONSTRAINT _follow_right_id_fkey;

ALTER TABLE _auth_role RENAME auth_id TO user_id;
ALTER TABLE _device RENAME auth_id TO user_id;
ALTER TABLE _subscription RENAME auth_id TO user_id;

ALTER TABLE _auth RENAME TO _user;

ALTER TABLE _auth_role RENAME TO _user_role;

ALTER TABLE _user_role
	ADD CONSTRAINT _user_role_user_id_fkey FOREIGN KEY (user_id) REFERENCES _user (id);
ALTER TABLE _device
	ADD CONSTRAINT _device_user_id_fkey FOREIGN KEY (user_id) REFERENCES _user (id);
ALTER TABLE _friend
	ADD CONSTRAINT _friend_right_id_fkey FOREIGN KEY (right_id) REFERENCES _user (id);
ALTER TABLE _follow
	ADD CONSTRAINT _follow_right_id_fkey FOREIGN KEY (right_id) REFERENCES _user (id);
	`

	_, err = tx.Exec(migrateSkygearUserStmt)
	return err
}
