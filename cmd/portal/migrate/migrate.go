package migrate

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
)

const migrationsDir = "migrations/portal"
const migrationsTable = "migration"

func init() {
	migrate.SetTable(migrationsTable)
}

func CreateMigration(name string) {
	const migrationTemplate = `-- +migrate Up

-- +migrate Down
`
	fileName := fmt.Sprintf("%s-%s.sql", time.Now().Format("20060102150405"), name)

	err := os.MkdirAll(migrationsDir, os.ModePerm)
	if err != nil {
		log.Fatalf("cannot ensure migrations directory: %s", err)
	}
	err = ioutil.WriteFile(path.Join(migrationsDir, fileName), []byte(migrationTemplate), os.ModePerm)
	if err != nil {
		log.Fatalf("cannot create new migration file: %s", err)
	}

	log.Printf("new migration '%s' created.", fileName)
}

type Options struct {
	DatabaseURL    string
	DatabaseSchema string
}

func Up(opts Options) {
	db := openDB(opts)
	migrations := &migrate.FileMigrationSource{
		Dir: migrationsDir,
	}

	n, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	log.Printf("applied %d migrations.", n)
	if err != nil {
		log.Fatalf("cannot apply all migrations: %s", err)
	}
}

func Down(numMigrations int, opts Options) {
	db := openDB(opts)
	migrations := &migrate.FileMigrationSource{
		Dir: migrationsDir,
	}

	n, err := migrate.ExecMax(db, "postgres", migrations, migrate.Down, numMigrations)
	log.Printf("reverted %d migrations.", n)
	if err != nil {
		log.Fatalf("cannot revert all migrations: %s", err)
	}
}

func Status(opts Options) bool {
	db := openDB(opts)
	migrations := &migrate.FileMigrationSource{
		Dir: migrationsDir,
	}

	plan, _, err := migrate.PlanMigration(db, "postgres", migrations, migrate.Up, 0)
	if err != nil {
		log.Fatalf("cannot plan migration: %s", err)
	}

	if len(plan) == 0 {
		log.Print("database schema is up-to-date")
	} else {
		log.Print("pending migrations:")
		for _, record := range plan {
			log.Printf("%s\n", record.Id)
		}
	}

	return len(plan) == 0
}

func openDB(opts Options) *sql.DB {
	db, err := sql.Open("postgres", opts.DatabaseURL)
	if err != nil {
		log.Fatalf("cannot open database: %s", err)
	}

	if opts.DatabaseSchema != "" {
		_, err = db.Exec(fmt.Sprintf("SET search_path TO %s", pq.QuoteIdentifier(opts.DatabaseSchema)))
		if err != nil {
			log.Fatalf("cannot set search_path: %s", err)
		}
	}

	return db
}
