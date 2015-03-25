package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/anachronistic/apns"

	"github.com/oursky/ourd/authtoken"
	"github.com/oursky/ourd/handler"
	"github.com/oursky/ourd/oddb"
	_ "github.com/oursky/ourd/oddb/fs"
	_ "github.com/oursky/ourd/oddb/pq"
	"github.com/oursky/ourd/push"
	"github.com/oursky/ourd/router"
	"github.com/oursky/ourd/subscription"
)

func usage() {
	fmt.Println("Usage: ourd [<config file>]")
}

func main() {
	var configPath string
	if len(os.Args) < 2 {
		configPath = os.Getenv("OD_CONFIG")
		if configPath == "" {
			usage()
			return
		}
	} else {
		configPath = os.Args[1]
	}

	config := Configuration{}
	if err := ReadFileInto(&config, configPath); err != nil {
		fmt.Println(err.Error())
		return
	}

	if config.Subscription.Enabled {
		pushSender := &push.APNSPusher{
			Client: apns.NewClient(config.APNS.Gateway, config.APNS.CertPath, config.APNS.KeyPath),
		}
		subscriptionService := &subscription.Service{
			ConnOpener:         func() (oddb.Conn, error) { return oddb.Open(config.DB.ImplName, config.App.Name, config.DB.Option) },
			NotificationSender: pushSender,
		}
		subscriptionService.Init()
	}

	naiveAPIKeyPreprocessor := apiKeyValidatonPreprocessor{
		Key: config.App.APIKey,
	}

	fileSystemConnPreprocessor := connPreprocessor{
		DBOpener: oddb.Open,
		DBImpl:   config.DB.ImplName,
		AppName:  config.App.Name,
		Option:   config.DB.Option,
	}

	fileTokenStorePreprocessor := tokenStorePreprocessor{
		Store: authtoken.FileStore(config.TokenStore.Path).Init(),
	}

	r := router.NewRouter()
	r.Map("", handler.HomeHandler)

	authPreprocessors := []router.Processor{
		naiveAPIKeyPreprocessor.Preprocess,
		fileSystemConnPreprocessor.Preprocess,
		fileTokenStorePreprocessor.Preprocess,
	}
	r.Map("auth:signup", handler.SignupHandler, authPreprocessors...)
	r.Map("auth:login", handler.LoginHandler, authPreprocessors...)

	recordPreprocessors := []router.Processor{
		naiveAPIKeyPreprocessor.Preprocess,
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
		naiveAPIKeyPreprocessor.Preprocess,
		fileSystemConnPreprocessor.Preprocess,
		fileTokenStorePreprocessor.Preprocess,
		authenticateUser,
	)

	// subscription shares the same set of preprocessor as record at the moment
	r.Map("subscription:save", handler.SubscriptionSaveHandler, recordPreprocessors...)

	log.Printf("Listening on %v...", config.HTTP.Host)
	err := http.ListenAndServe(config.HTTP.Host, r)
	if err != nil {
		log.Printf("Failed: %v", err)
	}
}
