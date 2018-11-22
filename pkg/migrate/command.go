package migrate

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/golang-migrate/migrate"
	log "github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/migrate/database/postgres"
)

func Run(module string, schema string, databaseURL string, sourceURL string, dryRun bool, command string, commandArg string) (err error) {
	if schema == "" {
		err = errors.New("missing schema")
		return
	}

	if module == "" {
		err = errors.New("missing module")
		return
	}

	if databaseURL == "" {
		err = errors.New("missing db url")
		return
	}

	if sourceURL == "" {
		err = errors.New("missing source url")
		return
	}

	versionTable := fmt.Sprintf("_%s_version", module)

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return
	}

	_, err = db.Exec(fmt.Sprintf("SET search_path TO %s", schema))
	if err != nil {
		return
	}

	config := postgres.Config{
		MigrationsTable: versionTable,
		DryRun:          dryRun,
	}
	driver, err := postgres.WithInstance(db, &config)
	if err != nil {
		return
	}
	defer driver.Close()

	m, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		return
	}

	err = runCommand(m, command, commandArg)
	return
}

func runCommand(m *migrate.Migrate, command string, commandArg string) (err error) {
	switch command {
	case "up":
		step, e := getStep(commandArg)
		if e != nil {
			err = e
			return
		}

		if step == -1 {
			err = m.Up()
		} else {
			err = m.Steps(step)
		}
	case "down":
		step, e := getStep(commandArg)
		if e != nil {
			err = e
			return
		}

		if step == -1 {
			err = m.Down()
		} else {
			err = m.Steps(-step)
		}
	case "force":
		v, e := strconv.ParseInt(commandArg, 10, 64)
		if e != nil {
			err = e
			return
		}

		err = m.Force(int(v))
	case "version":
		version, dirty, e := m.Version()
		if e != nil {
			err = e
			return
		}

		log.WithFields(log.Fields{
			"version": strconv.FormatInt(int64(version), 10),
			"dirty":   strconv.FormatBool(dirty),
		}).Info("checking version")
	default:
		err = errors.New("undefined command")
	}

	return
}

func getStep(stepStr string) (int, error) {
	if stepStr == "" {
		return -1, nil
	}

	n, err := strconv.ParseUint(stepStr, 10, 64)
	if err != nil {
		return -1, errors.New("invalid step")
	}

	return int(n), nil
}
