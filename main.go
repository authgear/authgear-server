package main

import (
	"log"
	"net/http"

	"github.com/oursky/ourd/authtoken"
	"github.com/oursky/ourd/handler"
	"github.com/oursky/ourd/oddb"
	_ "github.com/oursky/ourd/oddb/fs"
	"github.com/oursky/ourd/router"
)

func main() {
	config := Configuration{}
	ReadFileInto(&config, "development.ini")

	fileSystemConnPreprocessor := connPreprocessor{
		DBOpener: oddb.Open,
		DBImpl:   config.DB.ImplName,
		AppName:  config.DB.AppName,
		Option:   config.DB.Option,
	}

	fileTokenStorePreprocessor := tokenStorePreprocessor{
		Store: authtoken.FileStore(config.TokenStore.Path).Init(),
	}

	r := router.NewRouter()
	r.Map("", handler.HomeHandler)

	authPreprocessors := []router.Processor{
		fileSystemConnPreprocessor.Preprocess,
		fileTokenStorePreprocessor.Preprocess,
	}
	r.Map("auth:signup", handler.SignupHandler, authPreprocessors...)
	r.Map("auth:login", handler.LoginHandler, authPreprocessors...)

	recordPreprocessors := []router.Processor{
		fileSystemConnPreprocessor.Preprocess,
		fileTokenStorePreprocessor.Preprocess,
		authenticateUser,
		injectDatabase,
	}
	r.Map("record:fetch", handler.RecordFetchHandler, recordPreprocessors...)
	r.Map("record:query", handler.RecordQueryHandler, recordPreprocessors...)
	r.Map("record:save", handler.RecordSaveHandler, recordPreprocessors...)
	r.Map("record:delete", handler.RecordDeleteHandler, recordPreprocessors...)

	log.Printf("Listening on %v...", config.HTTP.Host)
	http.ListenAndServe(config.HTTP.Host, r)
}
