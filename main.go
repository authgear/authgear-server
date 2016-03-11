// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/evalphobia/logrus_sentry"
	"github.com/facebookgo/inject"
	"github.com/robfig/cron"

	"github.com/oursky/skygear/asset"
	"github.com/oursky/skygear/authtoken"
	"github.com/oursky/skygear/handler"
	"github.com/oursky/skygear/plugin"
	_ "github.com/oursky/skygear/plugin/exec"
	"github.com/oursky/skygear/plugin/hook"
	_ "github.com/oursky/skygear/plugin/http"
	"github.com/oursky/skygear/plugin/provider"
	_ "github.com/oursky/skygear/plugin/zmq"
	pp "github.com/oursky/skygear/preprocessor"
	"github.com/oursky/skygear/pubsub"
	"github.com/oursky/skygear/push"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skyconfig"
	"github.com/oursky/skygear/skydb"
	_ "github.com/oursky/skygear/skydb/pq"
	"github.com/oursky/skygear/subscription"
)

func usage() {
	fmt.Println("Usage: skygear [<config file>]")
}

func main() {
	var configPath string
	if len(os.Args) < 2 {
		configPath = os.Getenv("SKY_CONFIG")
		if configPath == "" {
			configPath = os.Getenv("OD_CONFIG")
			if configPath == "" {
				usage()
				return
			}
			fmt.Print("Config via OD_CONFIG will be deprecated in next version, use SKY_CONFIG\n")
		}
	} else {
		configPath = os.Args[1]
	}

	config := skyconfig.Configuration{}
	if err := skyconfig.ReadFileInto(&config, configPath); err != nil {
		fmt.Println(err.Error())
		return
	}

	initLogger(config)
	connOpener := ensureDB(config) // Fatal on DB failed

	// Init all the services
	r := router.NewRouter()
	serveMux := http.NewServeMux()
	pushSender := initPushSender(config, connOpener)

	tokenStore := authtoken.InitTokenStore(config.TokenStore.ImplName, config.TokenStore.Path)

	preprocessorRegistry := router.PreprocessorRegistry{}

	cronjob := cron.New()
	initContext := plugin.InitContext{
		Router:           r,
		Mux:              serveMux,
		Preprocessors:    preprocessorRegistry,
		HookRegistry:     hook.NewRegistry(),
		ProviderRegistry: provider.NewRegistry(),
		Scheduler:        cronjob,
		Config:           config,
	}

	internalHub := pubsub.NewHub()
	initSubscription(config, connOpener, internalHub, pushSender)
	initDevice(config, connOpener)

	// Preprocessor
	preprocessorRegistry["notification"] = &pp.NotificationPreprocessor{
		NotificationSender: pushSender,
	}
	preprocessorRegistry["accesskey"] = &pp.AccessKeyValidationPreprocessor{
		ClientKey: config.App.APIKey,
		MasterKey: config.App.MasterKey,
		AppName:   config.App.Name,
	}
	preprocessorRegistry["authenticator"] = &pp.UserAuthenticator{
		ClientKey:  config.App.APIKey,
		MasterKey:  config.App.MasterKey,
		AppName:    config.App.Name,
		TokenStore: tokenStore,
	}
	preprocessorRegistry["dbconn"] = &pp.ConnPreprocessor{
		AppName:       config.App.Name,
		AccessControl: config.App.AccessControl,
		DBOpener:      skydb.Open,
		DBImpl:        config.DB.ImplName,
		Option:        config.DB.Option,
	}
	preprocessorRegistry["plugin"] = &pp.EnsurePluginReadyPreprocessor{&initContext}
	preprocessorRegistry["inject_user"] = &pp.InjectUserIfPresent{}
	preprocessorRegistry["require_user"] = &pp.RequireUserForWrite{}
	preprocessorRegistry["inject_db"] = &pp.InjectDatabase{}
	preprocessorRegistry["inject_public_db"] = &pp.InjectPublicDatabase{}
	preprocessorRegistry["dev_only"] = &pp.DevOnlyProcessor{config.App.DevMode}

	r.Map("", &handler.HomeHandler{})

	g := &inject.Graph{}
	injectErr := g.Provide(
		&inject.Object{Value: initContext.ProviderRegistry, Complete: true, Name: "ProviderRegistry"},
		&inject.Object{Value: initContext.HookRegistry, Complete: true, Name: "HookRegistry"},
		&inject.Object{Value: tokenStore, Complete: true, Name: "TokenStore"},
		&inject.Object{Value: initAssetStore(config), Complete: true, Name: "AssetStore"},
		&inject.Object{Value: pushSender, Complete: true, Name: "PushSender"},
		&inject.Object{
			Value:    skydb.GetAccessModel(config.App.AccessControl),
			Complete: true,
			Name:     "AccessModel",
		},
	)
	if injectErr != nil {
		panic(fmt.Sprintf("Unable to set up handler: %v", injectErr))
	}

	injector := router.HandlerInjector{
		g,
		&preprocessorRegistry,
	}

	r.Map("auth:signup", injector.Inject(&handler.SignupHandler{}))
	r.Map("auth:login", injector.Inject(&handler.LoginHandler{}))
	r.Map("auth:logout", injector.Inject(&handler.LogoutHandler{}))
	r.Map("auth:password", injector.Inject(&handler.PasswordHandler{}))

	r.Map("record:fetch", injector.Inject(&handler.RecordFetchHandler{}))
	r.Map("record:query", injector.Inject(&handler.RecordQueryHandler{}))
	r.Map("record:save", injector.Inject(&handler.RecordSaveHandler{}))
	r.Map("record:delete", injector.Inject(&handler.RecordDeleteHandler{}))

	r.Map("device:register", injector.Inject(&handler.DeviceRegisterHandler{}))

	// subscription shares the same set of preprocessor as record read at the moment
	r.Map("subscription:fetch_all", injector.Inject(&handler.SubscriptionFetchAllHandler{}))
	r.Map("subscription:fetch", injector.Inject(&handler.SubscriptionFetchHandler{}))
	r.Map("subscription:save", injector.Inject(&handler.SubscriptionSaveHandler{}))
	r.Map("subscription:delete", injector.Inject(&handler.SubscriptionDeleteHandler{}))

	// relation shares the same setof preprocessor
	r.Map("relation:query", injector.Inject(&handler.RelationQueryHandler{}))
	r.Map("relation:add", injector.Inject(&handler.RelationAddHandler{}))
	r.Map("relation:remove", injector.Inject(&handler.RelationRemoveHandler{}))

	r.Map("user:query", injector.Inject(&handler.UserQueryHandler{}))
	r.Map("user:update", injector.Inject(&handler.UserUpdateHandler{}))
	r.Map("user:link", injector.Inject(&handler.UserLinkHandler{}))

	r.Map("role:default", injector.Inject(&handler.RoleDefaultHandler{}))
	r.Map("role:admin", injector.Inject(&handler.RoleAdminHandler{}))

	r.Map("push:user", injector.Inject(&handler.PushToUserHandler{}))
	r.Map("push:device", injector.Inject(&handler.PushToDeviceHandler{}))

	r.Map("schema:rename", injector.Inject(&handler.SchemaRenameHandler{}))
	r.Map("schema:delete", injector.Inject(&handler.SchemaDeleteHandler{}))
	r.Map("schema:create", injector.Inject(&handler.SchemaCreateHandler{}))
	r.Map("schema:fetch", injector.Inject(&handler.SchemaFetchHandler{}))
	r.Map("schema:access", injector.Inject(&handler.SchemaAccessHandler{}))

	serveMux.Handle("/", r)

	// Following section is for Gateway
	pubSub := pubsub.NewWsPubsub(nil)
	pubSubGateway := router.NewGateway("", "/pubsub", serveMux)
	pubSubGateway.GET(injector.InjectProcessors(&handler.PubSubHandler{
		WebSocket: pubSub,
	}))

	internalPubSub := pubsub.NewWsPubsub(internalHub)
	internalPubSubGateway := router.NewGateway("", "/_/pubsub", serveMux)
	internalPubSubGateway.GET(injector.InjectProcessors(&handler.PubSubHandler{
		WebSocket: internalPubSub,
	}))

	fileGateway := router.NewGateway("files/(.+)", "/files/", serveMux)
	fileGateway.GET(injector.Inject(&handler.AssetGetURLHandler{}))
	fileGateway.PUT(injector.Inject(&handler.AssetUploadURLHandler{}))

	corsHost := config.App.CORSHost

	var finalMux http.Handler
	if corsHost != "" {
		finalMux = &router.CORSMiddleware{
			Origin: corsHost,
			Next:   serveMux,
		}
	} else {
		finalMux = serveMux
	}

	if config.LOG.Level == "debug" {
		finalMux = &router.LoggingMiddleware{
			Skips: []string{
				"/files/",
				"/_/pubsub/",
				"/pubsub/",
			},
			Next: finalMux,
		}
	}

	// Bootstrap finished, starting services
	cronjob.Start()
	initPlugin(config, &initContext)

	log.Printf("Listening on %v...", config.HTTP.Host)
	err := http.ListenAndServe(config.HTTP.Host, finalMux)
	if err != nil {
		log.Printf("Failed: %v", err)
	}
}

func ensureDB(config skyconfig.Configuration) func() (skydb.Conn, error) {
	connOpener := func() (skydb.Conn, error) {
		return skydb.Open(
			config.DB.ImplName,
			config.App.Name,
			config.App.AccessControl,
			config.DB.Option,
			config.App.DevMode,
		)
	}

	// Attempt to open connection to database. Retry for a number of
	// times before giving up.
	attempt := 0
	for {
		conn, connError := connOpener()
		if connError == nil {
			conn.Close()
			return connOpener
		}

		attempt++
		log.Errorf("Failed to start skygear: %v", connError)
		if attempt >= 5 {
			log.Fatalf("Failed to start skygear because connection to database cannot be opened.")
		}

		log.Info("Retrying in 1 second...")
		time.Sleep(time.Second * time.Duration(1))
	}
}

func initAssetStore(config skyconfig.Configuration) asset.Store {
	var store asset.Store
	switch config.AssetStore.ImplName {
	default:
		panic("unrecgonized asset store implementation: " + config.AssetStore.ImplName)
	case "fs":
		store = asset.NewFileStore(
			config.AssetStore.Path,
			config.AssetURLSigner.URLPrefix,
			config.AssetURLSigner.Secret,
			config.AssetStore.Public,
		)
	case "s3":
		s3Store, err := asset.NewS3Store(
			config.AssetStore.AccessToken,
			config.AssetStore.SecretToken,
			config.AssetStore.Region,
			config.AssetStore.Bucket,
			config.AssetStore.Public,
		)
		if err != nil {
			panic("failed to initialize asset.S3Store: " + err.Error())
		}
		store = s3Store
	}
	return store
}

func initDevice(config skyconfig.Configuration, connOpener func() (skydb.Conn, error)) {
	// TODO: Create a device service to check APNs to remove obsolete devices.
	// The current implementaion deletes pubsub devices if the last registered
	// time is more than 1 day old.
	conn, err := connOpener()
	if err != nil {
		log.Warnf("Failed to delete outdated devices: %v", err)
	}

	conn.DeleteEmptyDevicesByTime(time.Now().AddDate(0, 0, -1))
}

func initPushSender(config skyconfig.Configuration, connOpener func() (skydb.Conn, error)) push.Sender {
	routeSender := push.NewRouteSender()
	if config.APNS.Enable {
		apns := initAPNSPusher(config, connOpener)
		routeSender.Route("aps", apns)
		routeSender.Route("ios", apns)
	}
	if config.GCM.Enable {
		gcm := initGCMPusher(config)
		routeSender.Route("gcm", gcm)
		routeSender.Route("android", gcm)
	}
	return routeSender
}

func initAPNSPusher(config skyconfig.Configuration, connOpener func() (skydb.Conn, error)) *push.APNSPusher {
	apnsPushSender, err := push.NewAPNSPusher(connOpener, push.GatewayType(config.APNS.Env), config.APNS.Cert, config.APNS.Key)
	if err != nil {
		log.Fatalf("Failed to set up push sender: %v", err)
	}
	go apnsPushSender.Run()
	go apnsPushSender.RunFeedback()

	return apnsPushSender
}

func initGCMPusher(config skyconfig.Configuration) *push.GCMPusher {
	return &push.GCMPusher{APIKey: config.GCM.APIKey}
}

func initSubscription(config skyconfig.Configuration, connOpener func() (skydb.Conn, error), hub *pubsub.Hub, pushSender push.Sender) {
	notifiers := []subscription.Notifier{subscription.NewHubNotifier(hub)}
	if pushSender != nil {
		notifiers = append(notifiers, subscription.NewPushNotifier(pushSender))
	}

	subscriptionService := &subscription.Service{
		ConnOpener: connOpener,
		Notifier:   subscription.NewMultiNotifier(notifiers...),
	}
	log.Infoln("Subscription Service listening...")
	go subscriptionService.Run()
}

func initPlugin(config skyconfig.Configuration, initContext *plugin.InitContext) {
	log.Infof("Supported plugin transports: %s", strings.Join(plugin.SupportedTransports(), ", "))

	for _, pluginConfig := range config.Plugin {
		initContext.AddPluginConfiguration(pluginConfig.Transport, pluginConfig.Path, pluginConfig.Args)
	}

	initContext.InitPlugins()
}

func initLogger(config skyconfig.Configuration) {
	// Setup Logging
	log.SetOutput(os.Stderr)
	level, err := log.ParseLevel(config.LOG.Level)
	if err != nil {
		log.Warnf("log: error parsing config: %v", err)
		log.Warnln("log: fall back to `debug`")
		level = log.DebugLevel
	}
	log.SetLevel(level)

	if config.LogHook.SentryDSN != "" {
		initSentry(config)
	}
}

func higherLogLevels(minLevel log.Level) []log.Level {
	levels := []log.Level{
		log.PanicLevel,
		log.FatalLevel,
		log.ErrorLevel,
		log.WarnLevel,
		log.InfoLevel,
		log.DebugLevel,
	}

	output := make([]log.Level, 0, len(levels))
	for _, level := range levels {
		if level <= minLevel {
			output = append(output, level)
		}
	}
	return output
}

func initSentry(config skyconfig.Configuration) {
	level, err := log.ParseLevel(config.LogHook.SentryLevel)
	if err != nil {
		log.Fatalf("log-hook: error parsing sentry-level: %v", err)
		return
	}

	levels := higherLogLevels(level)

	hook, err := logrus_sentry.NewSentryHook(config.LogHook.SentryDSN, levels)
	hook.Timeout = 1 * time.Second
	if err != nil {
		log.Errorf("Failed to initialize Sentry: %v", err)
		return
	}
	log.Infof("Logging to Sentry: %v", levels)
	log.AddHook(hook)
}
