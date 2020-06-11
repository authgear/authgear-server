package main

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

const migrationsDir = "migrate/migrations"
const migrationsTable = "_auth_migration"

func init() {
	migrate.SetTable(migrationsTable)
}

func cmdNew(args []string) {
	if len(args) == 0 {
		log.Fatal("must provide a name for new migration")
	}

	var name = strings.Join(args, "_")

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

func cmdUp(args []string) {
	db := openDB()
	migrations := &migrate.FileMigrationSource{
		Dir: migrationsDir,
	}

	n, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	log.Printf("applied %d migrations.", n)
	if err != nil {
		log.Fatalf("cannot apply all migrations: %s", err)
	}
}

func cmdDown(args []string) {
	db := openDB()
	migrations := &migrate.FileMigrationSource{
		Dir: migrationsDir,
	}

	n, err := migrate.Exec(db, "postgres", migrations, migrate.Down)
	log.Printf("reverted %d migrations.", n)
	if err != nil {
		log.Fatalf("cannot revert all migrations: %s", err)
	}
}

func cmdStatus(args []string) {
	db := openDB()

	records, err := migrate.GetMigrationRecords(db, "postgres")
	if err != nil {
		log.Fatalf("cannot get migration records: %s", err)
	}

	for _, record := range records {
		log.Printf("%s %s", record.AppliedAt.Format(time.RFC3339), record.Id)
	}
}

func openDB() *sql.DB {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("environment variable DATABASE_URL must be set")
	}
	dbSchema := os.Getenv("DATABASE_SCHEMA")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("cannot open database: %s", err)
	}

	if dbSchema != "" {
		_, err = db.Exec(fmt.Sprintf("SET search_path TO %s", pq.QuoteIdentifier(dbSchema)))
		if err != nil {
			log.Fatalf("cannot set search_path: %s", err)
		}
	}

	return db
}

var commands = map[string]func(args []string){
	"new":    cmdNew,
	"up":     cmdUp,
	"down":   cmdDown,
	"status": cmdStatus,
}

func usage() {
	log.Print("usage:")
	log.Print("  migrate new <name>")
	log.Print("  migrate up")
	log.Print("  migrate down")
	log.Print("  migrate status")
}

func main() {
	log.SetFlags(0)

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("failed to load .env: %s", err)
	}

	if len(os.Args) < 2 {
		log.Print("must provide sub-command")
		usage()
		os.Exit(1)
	}

	key := os.Args[1]
	cmd, ok := commands[key]
	if !ok {
		log.Printf("unrecognized sub-command: %s", key)
		usage()
		os.Exit(1)
	}

	cmd(os.Args[2:])
}
