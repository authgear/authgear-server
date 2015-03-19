package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/oursky/ourd/authtoken"
	"github.com/oursky/ourd/handler"
	"github.com/oursky/ourd/oddb"
	_ "github.com/oursky/ourd/oddb/fs"
	"github.com/oursky/ourd/router"
	"github.com/oursky/ourd/subscription"
)

func usage() {
	fmt.Println("Usage: ourd [<config file>]")
}

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}
	configPath := os.Args[1]

	config := Configuration{}
	if err := ReadFileInto(&config, configPath); err != nil {
		fmt.Println(err.Error())
		return
	}

	subscriptionService := &subscription.Service{
		ConnOpener: func() (oddb.Conn, error) { return oddb.Open(config.DB.ImplName, config.DB.AppName, config.DB.Option) },
	}
	subscriptionService.Init()

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

	r.Map("device:register",
		handler.DeviceRegisterHandler,
		fileSystemConnPreprocessor.Preprocess,
		fileTokenStorePreprocessor.Preprocess,
		authenticateUser,
	)

	r.Map("subscription:save", handler.SubscriptionSaveHandler, recordPreprocessors...)

	log.Printf("Listening on %v...", config.HTTP.Host)
	http.ListenAndServe(config.HTTP.Host, r)
}
