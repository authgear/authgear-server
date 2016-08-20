package migration

import "github.com/jmoiron/sqlx"

type revision_52b88c46931 struct {
}

func (r *revision_52b88c46931) Version() string { return "52b88c46931" }

func (r *revision_52b88c46931) Up(tx *sqlx.Tx) error {
	stmts := []string{
		`CREATE TABLE _record_creation (
		    record_type text NOT NULL,
		    role_id text,
		    UNIQUE (record_type, role_id),
		    FOREIGN KEY (role_id) REFERENCES _role(id)
		);`,
		`CREATE INDEX _record_creation_unique_record_type ON _record_creation (record_type);`,
	}

	for _, stmt := range stmts {
		_, err := tx.Exec(stmt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *revision_52b88c46931) Down(tx *sqlx.Tx) error {
	stmt := `DROP TABLE _schema;`
	_, err := tx.Exec(stmt)
	return err
}
