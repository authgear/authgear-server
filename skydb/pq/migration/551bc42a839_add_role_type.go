package migration

import (
	"github.com/jmoiron/sqlx"
)

type revision_551bc42a839 struct {
}

func (r *revision_551bc42a839) Version() string { return "551bc42a839" }

func (r *revision_551bc42a839) Up(tx *sqlx.Tx) error {
	stmts := []string{
		`ALTER TABLE _role ADD COLUMN by_default boolean DEFAULT FALSE;`,
		`ALTER TABLE _role ADD COLUMN is_admin boolean DEFAULT FALSE;`,
	}
	for _, stmt := range stmts {
		_, err := tx.Exec(stmt)
		return err
	}
	return nil
}

func (r *revision_551bc42a839) Down(tx *sqlx.Tx) error {
	stmts := []string{
		`ALTER TABLE _role DROP COLUMN is_admin;`,
		`ALTER TABLE _role DROP COLUMN by_default;`,
	}
	for _, stmt := range stmts {
		_, err := tx.Exec(stmt)
		return err
	}
	return nil
}
