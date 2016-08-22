package migration

import "github.com/jmoiron/sqlx"

type revision_c0397f15027 struct {
}

func (r *revision_c0397f15027) Version() string { return "c0397f15027" }

func (r *revision_c0397f15027) Up(tx *sqlx.Tx) error {
	stmts := []string{
		`CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA public;`,
		`ALTER TABLE _user ALTER COLUMN username TYPE citext;`,
		`ALTER TABLE _user ALTER COLUMN email TYPE citext;`,
	}

	for _, stmt := range stmts {
		_, err := tx.Exec(stmt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *revision_c0397f15027) Down(tx *sqlx.Tx) error {
	stmts := []string{
		`ALTER TABLE _user ALTER COLUMN username TYPE text;`,
		`ALTER TABLE _user ALTER COLUMN email TYPE text;`,
		`DROP EXTENSION citext;`,
	}

	for _, stmt := range stmts {
		_, err := tx.Exec(stmt)
		if err != nil {
			return err
		}
	}

	return nil
}
