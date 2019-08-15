package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	nurl "net/url"
	"os"

	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/lib/pq"
	"github.com/skygeario/skygear-server/migrate/pkg/migrate/command"
	"github.com/skygeario/skygear-server/migrate/pkg/migrate/source"
	"github.com/skygeario/skygear-server/migrate/pkg/migrate/tenant"
)

type commandRequest struct {
	Migration   string `json:"migration"`
	Schema      string `json:"schema"`
	DatabaseURL string `json:"database"`
	SourceURL   string `json:"-"`
	DryRun      bool   `json:"dry_run"`
	Command     string `json:"command"`
	CommandArg  string `json:"command_arg"`
}

func main() {
	// multi-tentant mode flags
	configDatabasePtr := flag.String("config-database", "", "(multi-tenant mode only) app config db url")
	appFilterKeyPtr := flag.String("app-filter-key", "", "filter app by key, e.g. auth_version")
	appFilterValuePtr := flag.String("app-filter-value", "", "use with app-filter-key, specify the value of key, e.g. live")
	hostnameOverridePtr := flag.String("hostname-override", "", "override the hostname of the database url of each app")

	// single db mode flags
	databasePtr := flag.String("database", "postgres://postgres:@localhost/postgres?sslmode=disable", "migration db url")
	schemaPtr := flag.String("schema", "", "migration schema")

	// general flags
	sources := command.SourceFlags{}
	flag.Var(&sources, "add-migration-src", "specify source in form <migration_name>,<source_url>")

	migrations := command.MigrationFlags{}
	flag.Var(&migrations, "migration", "migration name, e.g. core")

	dryRunPtr := flag.Bool("dry-run", false, "enable dry run will rollback the transaction")
	startHTTPServerPtr := flag.Bool("http-server", false, "start server to accept migration request by api. if this is true, all the other flag will be ignored")

	flag.Parse()

	configDBStr := *configDatabasePtr
	appFilterKey := *appFilterKeyPtr
	appFilterValue := *appFilterValuePtr
	hostnameOverride := *hostnameOverridePtr
	schema := *schemaPtr
	databaseURL := *databasePtr
	dryRun := *dryRunPtr
	command := flag.Arg(0)
	commandArg := flag.Arg(1)

	// Download source
	err := source.ClearCache()
	if err != nil {
		log.WithField("error", err).Error("unable to clear cache")
		os.Exit(1)
	}

	for _, v := range sources {
		newSourceURL, err := source.Download(v.SourceURL)
		if err != nil {
			log.WithField("error", err).Error("unable to download source")
			os.Exit(1)
		}
		(*v).SourceURL = newSourceURL
	}

	startHTTPServer := *startHTTPServerPtr
	if !startHTTPServer {
		if configDBStr != "" {
			// multi-tenant mode
			runMultiTenantMigrations(
				configDBStr, appFilterKey, appFilterValue,
				hostnameOverride, sources, migrations,
				dryRun, command, commandArg,
			)
		} else {
			// cli mode
			for _, m := range migrations {
				SourceURL := ""
				if s, ok := sources[m]; ok {
					SourceURL = s.SourceURL
				}
				_, err := runCmd(commandRequest{
					Migration:   m,
					Schema:      schema,
					DatabaseURL: databaseURL,
					SourceURL:   SourceURL,
					DryRun:      dryRun,
					Command:     command,
					CommandArg:  commandArg,
				})
				if err != nil {
					os.Exit(1)
				}
			}
		}
	} else {
		// http mode
		http.HandleFunc("/migrate", func(w http.ResponseWriter, r *http.Request) {
			var err error
			var result string
			var payload commandRequest
			if r.Body == nil {
				http.Error(w, "Please send a request body", 400)
				return
			}
			err = json.NewDecoder(r.Body).Decode(&payload)
			defer func() {
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error": err.Error(),
					})
				} else {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"result": result,
					})
				}
			}()
			sourceURL, ok := sources[payload.Migration]
			if !ok {
				err = fmt.Errorf("unknown migration: %v", payload.Migration)
				return
			}
			payload.SourceURL = sourceURL.SourceURL
			result, err = runCmd(payload)
		})
		http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"result": "OK",
			})
		})
		log.Printf("migration server boot")
		log.Fatal(http.ListenAndServe(":3000", nil))
	}
}

func runMultiTenantMigrations(
	configDBStr string, appFilterKey string, appFilterValue string,
	hostnameOverride string, sources command.SourceFlags,
	migrations command.MigrationFlags,
	dryRun bool, command string, commandArg string,
) {
	db, err := sql.Open("postgres", configDBStr)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer func() {
		err = db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	apps, err := tenant.GetMigrateApps(hostnameOverride, db, appFilterKey, appFilterValue)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.WithFields(log.Fields{
		"count":        len(apps),
		"filter_key":   appFilterKey,
		"filter_value": appFilterValue,
	}).Info("start running migration")

	for _, a := range apps {
		for _, migration := range migrations {
			databaseURL := a.DatabaseURL
			schema := fmt.Sprintf("app_%s", a.Name)

			sourceURL := ""
			if s, ok := sources[migration]; ok {
				sourceURL = s.SourceURL
			}

			_, err = runCmd(commandRequest{
				Migration:   migration,
				Schema:      schema,
				DatabaseURL: databaseURL,
				SourceURL:   sourceURL,
				DryRun:      dryRun,
				Command:     command,
				CommandArg:  commandArg,
			})

			if err != nil {
				os.Exit(1)
			}
		}
	}
}

func runCmd(req commandRequest) (string, error) {
	sourceURL := req.SourceURL
	surl, _ := nurl.Parse(sourceURL)
	if surl.Scheme == "" {
		sourceURL = fmt.Sprintf("file://%s", sourceURL)
	}

	schema := req.Schema
	if req.Migration == "gateway" {
		schema = "app_config"
	}

	purl, _ := nurl.Parse(req.DatabaseURL)
	l := log.WithFields(log.Fields{
		"migration":  req.Migration,
		"db_name":    purl.EscapedPath(),
		"db_host":    purl.Hostname(),
		"source_url": sourceURL,
		"schema":     schema,
	})

	result, err := command.Run(
		req.Migration,
		schema,
		req.DatabaseURL,
		sourceURL,
		req.DryRun,
		req.Command,
		req.CommandArg,
	)

	if err != nil {
		l.WithField("error", err).Error(err.Error())
	} else {
		l.WithField("result", result).Info("done")
	}

	return result, err
}
