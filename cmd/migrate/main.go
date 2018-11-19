package main

import (
	"database/sql"
	"flag"
	"fmt"
	"path"
	"strconv"

	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/lib/pq"

	"github.com/skygeario/skygear-server/pkg/migrate/database/postgres"
)

func main() {
	var err error

	modulePtr := flag.String("module", "", "module name, e.g. gateway, core, auth, record")
	databasePtr := flag.String("database", "postgres://postgres:@localhost/postgres?sslmode=disable", "migration db url")
	schemaPtr := flag.String("schema", "app__", "migration schema")
	dirPtr := flag.String("dir", "cmd/migrate/revisions", "(optional) directory of revisions files")
	dryRunPtr := flag.Bool("dry-run", false, "enable dry run will rollback the transaction")

	flag.Parse()

	schema := *schemaPtr
	if *modulePtr == "gateway" {
		schema = "app_config"
	}

	if schema == "" {
		panic("missing schema")
	}

	if *modulePtr == "" {
		panic("missing module")
	}

	filePath := fmt.Sprintf("file://%s", path.Join(*dirPtr, *modulePtr))
	versionTable := fmt.Sprintf("_%s_version", *modulePtr)

	db, err := sql.Open("postgres", *databasePtr)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(fmt.Sprintf("SET search_path TO %s", schema))
	if err != nil {
		panic(err)
	}

	config := postgres.Config{
		MigrationsTable: versionTable,
		DryRun:          *dryRunPtr,
	}
	driver, err := postgres.WithInstance(db, &config)
	if err != nil {
		panic(err)
	}
	defer driver.Close()

	m, err := migrate.NewWithDatabaseInstance(filePath, "postgres", driver)
	if err != nil {
		panic(err)
	}

	fmt.Println("Path: " + filePath)
	fmt.Println("Module namespace: " + *modulePtr)
	fmt.Println("Database: " + *databasePtr)
	fmt.Println("Schema: " + schema)

	err = runCommand(m)
	if err != nil {
		panic(err)
	}
}

func runCommand(m *migrate.Migrate) (err error) {
	switch flag.Arg(0) {
	case "up":
		step := getStep()

		if step == -1 {
			err = m.Up()
		} else {
			err = m.Steps(step)
		}
	case "down":
		step := getStep()

		if step == -1 {
			err = m.Down()
		} else {
			err = m.Steps(-step)
		}
	case "force":
		v, e := strconv.ParseInt(flag.Arg(1), 10, 64)
		if e != nil {
			panic(e)
		}

		err = m.Force(int(v))
	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			panic(err)
		}

		fmt.Println("Version: " + strconv.FormatInt(int64(version), 10))
		fmt.Println("Dirty: " + strconv.FormatBool(dirty))
	default:
		panic("Undefined command")
	}

	return
}

func getStep() int {
	if flag.Arg(1) == "" {
		return -1
	}

	n, err := strconv.ParseUint(flag.Arg(1), 10, 64)
	if err != nil {
		panic(err)
	}

	return int(n)
}
