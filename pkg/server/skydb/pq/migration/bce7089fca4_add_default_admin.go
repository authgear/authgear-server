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
	"golang.org/x/crypto/bcrypt"

	"github.com/skygeario/skygear-server/pkg/server/uuid"
)

const adminRoleDefaultName = "Admin"
const adminUserDefaultUsername = "admin"
const adminUserDefaultPassword = "secret"

type revision_bce7089fca4 struct {
}

func (r *revision_bce7089fca4) Version() string { return "bce7089fca4" }

func (r *revision_bce7089fca4) insertAdminRoleIfNotExists(tx *sqlx.Tx) error {
	var exists = false
	var err error

	if err = tx.QueryRowx(
		`SELECT EXISTS (SELECT 1 FROM _role WHERE is_admin = TRUE)`,
	).Scan(&exists); err != nil {
		return err
	}

	if exists {
		return nil
	}

	_, err = tx.Exec(
		`INSERT INTO _role (id, is_admin) VALUES ($1, TRUE)`,
		adminRoleDefaultName,
	)

	return err
}

func (r *revision_bce7089fca4) insertAdminUserIfNotExists(tx *sqlx.Tx) error {
	var err error
	var exists = false

	if err = tx.QueryRowx(`
		SELECT EXISTS (
			SELECT 1
			FROM _user_role as ur JOIN _role as r ON ur.role_id = r.id
			WHERE r.is_admin = TRUE
		)`).Scan(&exists); err != nil {
		return err
	}

	if exists {
		return nil
	}

	newUserID := uuid.New()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminUserDefaultPassword), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	if _, err = tx.Exec(`
		INSERT INTO _user (id, username, password)
		VALUES ($1, $2, $3)
	`, newUserID, adminUserDefaultUsername, hashedPassword); err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO _user_role (user_id, role_id)
		VALUES(
			$1,
			(
				SELECT id
				FROM _role
				WHERE _role.is_admin = TRUE
				LIMIT 1
			)
		)`, newUserID)

	return err
}

func (r *revision_bce7089fca4) Up(tx *sqlx.Tx) error {
	if err := r.insertAdminRoleIfNotExists(tx); err != nil {
		return err
	}

	return r.insertAdminUserIfNotExists(tx)
}

func (r *revision_bce7089fca4) Down(tx *sqlx.Tx) error {
	// nothing can be done for downgrade
	return nil
}
