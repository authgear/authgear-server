package main

import (
	"encoding/json"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	nurl "net/url"
	"os"
	"path"

	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/lib/pq"
	migrateCommand "github.com/skygeario/skygear-server/pkg/migrate"
)

type commandRequest struct {
	Module      string `json:"module"`
	Schema      string `json:"schema"`
	DatabaseURL string `json:"database"`
	RevDir      string `json:"dir"`
	DryRun      bool   `json:"dry_run"`
	Command     string `json:"command"`
	CommandArg  string `json:"command_arg"`
}

func main() {
	modulePtr := flag.String("module", "", "module name, e.g. gateway, core, auth, record")
	databasePtr := flag.String("database", "postgres://postgres:@localhost/postgres?sslmode=disable", "migration db url")
	schemaPtr := flag.String("schema", "", "migration schema")
	dirPtr := flag.String("dir", "", "(optional) directory of revisions files")
	dryRunPtr := flag.Bool("dry-run", false, "enable dry run will rollback the transaction")
	startHTTPServerPtr := flag.Bool("http-server", false, "start server to accept migration request by api. if this is true, all the other flag will be ignored")

	flag.Parse()

	module := *modulePtr
	schema := *schemaPtr
	databaseURL := *databasePtr
	revDir := *dirPtr
	dryRun := *dryRunPtr
	command := flag.Arg(0)
	commandArg := flag.Arg(1)

	startHTTPServer := *startHTTPServerPtr

	if !startHTTPServer {
		_, err := runCmd(commandRequest{
			Module:      module,
			Schema:      schema,
			DatabaseURL: databaseURL,
			RevDir:      revDir,
			DryRun:      dryRun,
			Command:     command,
			CommandArg:  commandArg,
		})
		if err != nil {
			os.Exit(1)
		}
	} else {
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
			result, err = runCmd(payload)
		})
		http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"result": "OK",
			})
		})
		log.Fatal(http.ListenAndServe(":3000", nil))
	}
}

func runCmd(req commandRequest) (string, error) {
	revDir := req.RevDir
	if revDir == "" {
		revDir = "cmd/migrate/revisions"
	}
	filePath := fmt.Sprintf("file://%s", path.Join(revDir, req.Module))

	schema := req.Schema
	if req.Module == "gateway" {
		schema = "app_config"
	}

	purl, _ := nurl.Parse(req.DatabaseURL)
	l := log.WithFields(log.Fields{
		"module":  req.Module,
		"db_name": purl.EscapedPath(),
		"db_host": purl.Hostname(),
		"schema":  schema,
	})

	result, err := migrateCommand.Run(
		req.Module,
		schema,
		req.DatabaseURL,
		filePath,
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
