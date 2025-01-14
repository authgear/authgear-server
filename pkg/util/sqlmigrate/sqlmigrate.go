package sqlmigrate

import (
	"database/sql"
	"embed"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/lib/pq"
	"github.com/rubenv/sql-migrate"
)

type ConnectionOptions struct {
	DatabaseURL    string
	DatabaseSchema string
}

type MigrationSet struct {
	MigrationSet                         migrate.MigrationSet
	EmbedFS                              embed.FS
	EmbedFSRoot                          string
	OutputPathRelativeToWorkingDirectory string
}

type NewMigrationSetOptions struct {
	// TableName is the name of the table that stores migration status.
	TableName string

	// EmbedFS is a embed.FS that stores migration files.
	// You also pass in a correct EmbedFSRoot to tell where to find the migrations.
	// Suppose you write //go:embed migrations/authgear
	// then EmbedFSRoot MUST BE "migrations/authgear" (no leading slash nor trailing slash)
	EmbedFS     embed.FS
	EmbedFSRoot string

	// OutputPathRelativeToWorkingDirectory is for the create command.
	// The create command create new file in the actual filesystem, so
	// it is relative to workdir.
	OutputPathRelativeToWorkingDirectory string
}

func NewMigrateSet(options NewMigrationSetOptions) MigrationSet {
	return MigrationSet{
		MigrationSet: migrate.MigrationSet{
			TableName: options.TableName,
		},
		EmbedFS:                              options.EmbedFS,
		EmbedFSRoot:                          options.EmbedFSRoot,
		OutputPathRelativeToWorkingDirectory: options.OutputPathRelativeToWorkingDirectory,
	}
}

func (s MigrationSet) Create(name string) (fileName string, err error) {
	const migrationTemplate = `-- +migrate Up

-- +migrate Down
`
	fileName = fmt.Sprintf("%s-%s.sql", time.Now().Format("20060102150405"), name)

	err = os.MkdirAll(s.OutputPathRelativeToWorkingDirectory, os.ModePerm)
	if err != nil {
		return
	}

	err = ioutil.WriteFile(path.Join(s.OutputPathRelativeToWorkingDirectory, fileName), []byte(migrationTemplate), 0o600)
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
		OriginSource: &migrate.EmbedFileSystemMigrationSource{
			FileSystem: s.EmbedFS,
			Root:       s.EmbedFSRoot,
		},
		Data: map[string]interface{}{
			"SCHEMA": opts.DatabaseSchema,
		},
	}
}
