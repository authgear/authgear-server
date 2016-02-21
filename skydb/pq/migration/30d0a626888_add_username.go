package migration

import (
	"github.com/jmoiron/sqlx"
)

type revision_30d0a626888 struct {
}

func (r *revision_30d0a626888) Version() string { return "30d0a626888" }

func (r *revision_30d0a626888) Up(tx *sqlx.Tx) error {
	stmts := []string{
		`ALTER TABLE _user ADD COLUMN username VARCHAR(255);`,
		`ALTER TABLE _user ADD UNIQUE (username);`,
		`ALTER TABLE _user ADD CONSTRAINT _user_email_key UNIQUE (email);`,
	}
	for _, stmt := range stmts {
		_, err := tx.Exec(stmt)
		return err
	}
	return nil
}

func (r *revision_30d0a626888) Down(tx *sqlx.Tx) error {
	stmts := []string{
		`ALTER TABLE _user DROP COLUMN username;`,
		`ALTER TABLE _user DROP CONSTRAINT _user_email_key;`,
	}
	for _, stmt := range stmts {
		_, err := tx.Exec(stmt)
		return err
	}
	return nil
}
