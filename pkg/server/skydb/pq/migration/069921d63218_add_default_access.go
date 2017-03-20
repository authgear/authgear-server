package migration

import "github.com/jmoiron/sqlx"

type revision_069921d63218 struct {
}

func (r *revision_069921d63218) Version() string { return "069921d63218" }

func (r *revision_069921d63218) Up(tx *sqlx.Tx) error {
	stmts := []string{
		`CREATE TABLE _record_default_access (
		    record_type text NOT NULL,
		    default_access jsonb,
		    UNIQUE (record_type)
		);`,
		`CREATE INDEX _record_default_access_unique_record_type ON _record_default_access (record_type);`,
	}

	for _, stmt := range stmts {
		_, err := tx.Exec(stmt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *revision_069921d63218) Down(tx *sqlx.Tx) error {
	stmt := `DROP TABLE _record_default_access;`
	_, err := tx.Exec(stmt)
	return err
}
