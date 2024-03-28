package e2e

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/lib/pq"
)

func openDB(dbURL string, dbSchema string) *sql.DB {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to open database: %s", err)
	}

	_, err = db.Exec(fmt.Sprintf("SET search_path TO %s", pq.QuoteIdentifier(dbSchema)))
	if err != nil {
		log.Fatalf("failed to set search_path: %s", err)
	}

	return db
}
