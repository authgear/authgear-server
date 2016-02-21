package migration

import (
	"github.com/jmoiron/sqlx"
)

type revision_51375067b45 struct {
}

func (r *revision_51375067b45) Version() string { return "51375067b45" }

func (r *revision_51375067b45) Up(tx *sqlx.Tx) error {
	_, err := tx.Exec(`ALTER TABLE _device ALTER COLUMN token DROP NOT NULL;`)
	return err
}

func (r *revision_51375067b45) Down(tx *sqlx.Tx) error {
	_, err := tx.Exec(`ALTER TABLE _device ALTER COLUMN token SET NOT NULL;`)
	return err
}
