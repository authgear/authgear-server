package sqlmigrate

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/lib/pq"
	"github.com/rubenv/sql-migrate"
)

type MigrationSet struct {
	MigrationSet migrate.MigrationSet
	Dir          string
}

type ConnectionOptions struct {
	DatabaseURL    string
	DatabaseSchema string
}

func NewMigrateSet(tableName string, dir string) MigrationSet {
	return MigrationSet{
		MigrationSet: migrate.MigrationSet{
			TableName: tableName,
		},
		Dir: dir,
	}
}

func (s MigrationSet) Create(name string) (fileName string, err error) {
	const migrationTemplate = `-- +migrate Up

-- +migrate Down
`
	fileName = fmt.Sprintf("%s-%s.sql", time.Now().Format("20060102150405"), name)

	err = os.MkdirAll(s.Dir, os.ModePerm)
	if err != nil {
		return
	}

	err = ioutil.WriteFile(path.Join(s.Dir, fileName), []byte(migrationTemplate), os.ModePerm)
	if err != nil {
		return
	}

	log.Printf("new migration '%s' created.", fileName)

	return
}

func (s MigrationSet) Up(opts ConnectionOptions, max int) (n int, err error) {
	db, err := s.openDB(opts)
	if err != nil {
		return
	}

	source := s.makeSource(opts)

	n, err = s.MigrationSet.ExecMax(db, "postgres", source, migrate.Up, max)
	if err != nil {
		return
	}

	log.Printf("applied %d migrations.", n)
	return
}

func (s MigrationSet) Down(opts ConnectionOptions, max int) (n int, err error) {
	db, err := s.openDB(opts)
	if err != nil {
		return
	}

	source := s.makeSource(opts)

	n, err = s.MigrationSet.ExecMax(db, "postgres", source, migrate.Down, max)
	if err != nil {
		return
	}

	log.Printf("reverted %d migrations.", n)
	return
}

func (s MigrationSet) Status(opts ConnectionOptions) (plans []*migrate.PlannedMigration, err error) {
	db, err := s.openDB(opts)
	if err != nil {
		return
	}

	source := s.makeSource(opts)

	plans, _, err = s.MigrationSet.PlanMigration(db, "postgres", source, migrate.Up, 0)
	if err != nil {
		return
	}

	if len(plans) == 0 {
		log.Print("database schema is up-to-date")
	} else {
		log.Print("pending migrations:")
		for _, plan := range plans {
			log.Printf("%s\n", plan.Id)
		}
	}

	return
}

func (s MigrationSet) openDB(opts ConnectionOptions) (db *sql.DB, err error) {
	db, err = sql.Open("postgres", opts.DatabaseURL)
	if err != nil {
		return
	}

	if opts.DatabaseSchema != "" {
		_, err = db.Exec(fmt.Sprintf("SET search_path TO %s", pq.QuoteIdentifier(opts.DatabaseSchema)))
		if err != nil {
			return
		}
	}

	return
}

func (s MigrationSet) makeSource(opts ConnectionOptions) *TemplateMigrationSource {
	return &TemplateMigrationSource{
		OriginSource: &migrate.FileMigrationSource{
			Dir: s.Dir,
		},
		Data: map[string]interface{}{
			"SCHEMA": opts.DatabaseSchema,
		},
	}
}
