package main

import (
	"database/sql"
	"flag"
	"fmt"
	"strconv"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/lib/pq"
)

func main() {
	var err error

	pathPtr := flag.String("path", "", "")
	gearPtr := flag.String("gear", "", "")
	databasePtr := flag.String("database", "postgres://postgres:@localhost/postgres?sslmode=disable", "")
	schemaPtr := flag.String("schema", "app__", "")

	flag.Parse()

	db, err := sql.Open("postgres", *databasePtr)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(fmt.Sprintf("SET search_path TO %s", *schemaPtr))
	if err != nil {
		panic(err)
	}

	config := postgres.Config{
		MigrationsTable: fmt.Sprintf("_%s_version", *gearPtr),
	}
	driver, err := postgres.WithInstance(db, &config)
	if err != nil {
		panic(err)
	}

	path := fmt.Sprintf("file://%s", *pathPtr)
	m, err := migrate.NewWithDatabaseInstance(path, "postgres", driver)
	if err != nil {
		panic(err)
	}

	fmt.Println("Path: " + *pathPtr)
	fmt.Println("Gear: " + *gearPtr)
	fmt.Println("Database: " + *databasePtr)
	fmt.Println("Schema: " + *schemaPtr)

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
		v, err := strconv.ParseInt(flag.Arg(1), 10, 64)
		if err != nil {
			panic(err)
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
