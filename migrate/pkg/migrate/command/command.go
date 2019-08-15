package command

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/golang-migrate/migrate"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/migrate/pkg/migrate/database/postgres"
)

func Run(module string, schema string, databaseURL string, sourceURL string, dryRun bool, command string, commandArg string) (result string, err error) {
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

	_, err = db.Exec(fmt.Sprintf("SET search_path TO %s", pq.QuoteIdentifier(schema)))
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

	result, err = runCommand(m, command, commandArg)
	return
}

func runCommand(m *migrate.Migrate, command string, commandArg string) (result string, err error) {
	switch command {
	case "up":
		step, e := getStep(commandArg)
		if e != nil {
			err = e
			return
		}

		result, err = upCmd(m, step)
	case "down":
		step, e := getStep(commandArg)
		if e != nil {
			err = e
			return
		}

		result, err = downCmd(m, step)
	case "force":
		v, e := strconv.ParseInt(commandArg, 10, 64)
		if e != nil {
			err = e
			return
		}

		result, err = forceCmd(m, int(v))
	case "version":
		version, dirty, e := m.Version()
		if e != nil {
			err = e
			return
		}

		result = fmt.Sprintf("%d", version)

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

func upCmd(m *migrate.Migrate, step int) (result string, err error) {
	var e error
	if step == -1 {
		e = m.Up()
	} else {
		e = m.Steps(step)
	}

	if e == nil {
		result = "ok"
	} else if e == migrate.ErrNoChange {
		result = "no change"
	} else {
		err = e
	}
	return
}

func downCmd(m *migrate.Migrate, step int) (result string, err error) {
	var e error
	if step == -1 {
		e = m.Down()
	} else {
		e = m.Steps(-step)
	}

	if e == nil {
		result = "ok"
	} else if e == migrate.ErrNoChange {
		result = "no change"
	} else {
		err = e
	}
	return
}

func forceCmd(m *migrate.Migrate, v int) (result string, err error) {
	if err = m.Force(v); err != nil {
		return
	}
	result = "ok"
	return
}
