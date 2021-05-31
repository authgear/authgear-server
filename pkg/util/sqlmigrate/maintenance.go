package sqlmigrate

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

type PartmanMaintainer struct {
	DatabaseURL    string
	DatabaseSchema string
	TableName      string
}

func (m PartmanMaintainer) RunMaintenance() (err error) {
	db, err := m.openDB()
	if err != nil {
		return
	}

	_, err = db.Exec(fmt.Sprintf("CALL partition_data_proc('%s.%s');", m.DatabaseSchema, m.TableName))
	if err != nil {
		return
	}

	_, err = db.Exec("CALL run_maintenance_proc();")
	if err != nil {
		return
	}

	return
}

func (m PartmanMaintainer) openDB() (db *sql.DB, err error) {
	db, err = sql.Open("postgres", m.DatabaseURL)
	if err != nil {
		return
	}

	if m.DatabaseSchema != "" {
		_, err = db.Exec(fmt.Sprintf("SET search_path TO %s", pq.QuoteIdentifier(m.DatabaseSchema)))
		if err != nil {
			return
		}
	}

	return
}
