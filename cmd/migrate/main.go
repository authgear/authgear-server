package main

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	nurl "net/url"
	"os"
	"path"

	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/lib/pq"
	migrateCommand "github.com/skygeario/skygear-server/pkg/migrate"
)

func main() {
	var err error

	modulePtr := flag.String("module", "", "module name, e.g. gateway, core, auth, record")
	databasePtr := flag.String("database", "postgres://postgres:@localhost/postgres?sslmode=disable", "migration db url")
	schemaPtr := flag.String("schema", "app__", "migration schema")
	dirPtr := flag.String("dir", "cmd/migrate/revisions", "(optional) directory of revisions files")
	dryRunPtr := flag.Bool("dry-run", false, "enable dry run will rollback the transaction")

	flag.Parse()

	module := *modulePtr
	schema := *schemaPtr
	databaseURL := *databasePtr
	dryRun := *dryRunPtr
	filePath := fmt.Sprintf("file://%s", path.Join(*dirPtr, *modulePtr))

	if module == "gateway" {
		schema = "app_config"
	}

	purl, _ := nurl.Parse(databaseURL)
	l := log.WithFields(log.Fields{
		"module":  module,
		"db_name": purl.EscapedPath(),
		"db_host": purl.Hostname(),
		"schema":  schema,
	})

	err = migrateCommand.Run(
		module,
		schema,
		databaseURL,
		filePath,
		dryRun,
		flag.Arg(0),
		flag.Arg(1),
	)

	if err != nil {
		l.WithField("error", err).Error(err.Error())
		os.Exit(1)
	}

	l.Info("done")
}
