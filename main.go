package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/Sirupsen/logrus/hooks/sentry"
	"github.com/facebookgo/inject"
	"github.com/robfig/cron"

	"github.com/oursky/skygear/asset"
	"github.com/oursky/skygear/authtoken"
	"github.com/oursky/skygear/handler"
	"github.com/oursky/skygear/plugin"
	_ "github.com/oursky/skygear/plugin/exec"
	"github.com/oursky/skygear/plugin/hook"
	"github.com/oursky/skygear/plugin/provider"
	_ "github.com/oursky/skygear/plugin/zmq"
	pp "github.com/oursky/skygear/preprocessor"
	"github.com/oursky/skygear/pubsub"
	"github.com/oursky/skygear/push"
	"github.com/oursky/skygear/router"
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

	initLogger(config)
	connOpener := ensureDB(config) // Fatal on DB failed

	// Init all the services
	r := router.NewRouter()
	pushSender := initPushSender(config, connOpener)

	tokenStore := authtoken.InitTokenStore(config.TokenStore.ImplName, config.TokenStore.Path)

	c := cron.New()
	initContext := plugin.InitContext{
		Router:           r,
		HookRegistry:     hook.NewRegistry(),
		ProviderRegistry: provider.NewRegistry(),
		Scheduler:        c,
	}
	c.Start()
	initPlugin(config, &initContext) // Block until plugin configured

	internalHub := pubsub.NewHub()
	initSubscription(config, connOpener, internalHub, pushSender)
	initDevice(config, connOpener)

	// Preprocessor
	notificationPreprocessor := pp.NotificationPreprocessor{
		NotificationSender: pushSender,
	}

	naiveAPIKeyPreprocessor := pp.AccessKeyValidatonPreprocessor{
		Key:     config.App.APIKey,
		AppName: config.App.Name,
	}

	authenticator := pp.UserAuthenticator{
		APIKey:     config.App.APIKey,
		AppName:    config.App.Name,
		TokenStore: tokenStore,
	}

	dbConnPreprocessor := pp.ConnPreprocessor{
		AppName:       config.App.Name,
		AccessControl: config.App.AccessControl,
		DBOpener:      skydb.Open,
		DBImpl:        config.DB.ImplName,
		Option:        config.DB.Option,
	}

	pluginReadyPreprocessor := &pp.EnsurePluginReadyPreprocessor{&initContext}

	baseAuthPreprocessors := []router.Processor{
		authenticator.Preprocess,
		dbConnPreprocessor.Preprocess,
		pp.InjectUserIfPresent,
		pp.InjectDatabase,
	}

	assetGetPreprocessors := []router.Processor{
		dbConnPreprocessor.Preprocess,
	}

	assetPutPreprocessors := []router.Processor{
		naiveAPIKeyPreprocessor.Preprocess,
		dbConnPreprocessor.Preprocess,
	}

	authPreprocessors := []router.Processor{
		naiveAPIKeyPreprocessor.Preprocess,
		dbConnPreprocessor.Preprocess,
		pluginReadyPreprocessor.Preprocess,
	}

	recordWritePreprocessors := append(baseAuthPreprocessors,
		pp.RequireUserForWrite,
		pluginReadyPreprocessor.Preprocess,
	)

	requireUserWritePreprocessors := append(baseAuthPreprocessors,
		pp.RequireUserForWrite,
	)

	notificationPreprocessors := []router.Processor{
		naiveAPIKeyPreprocessor.Preprocess,
		dbConnPreprocessor.Preprocess,
		pp.InjectDatabase,
		notificationPreprocessor.Preprocess,
	}

	pubSubPreprocessors := []router.Processor{
		naiveAPIKeyPreprocessor.Preprocess,
	}

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

	r.Map("auth:signup", injectedHandler(g, &handler.SignupHandler{}), authPreprocessors...)
	r.Map("auth:login", injectedHandler(g, &handler.LoginHandler{}), authPreprocessors...)
	r.Map("auth:logout", injectedHandler(g, &handler.LogoutHandler{}),
		authenticator.Preprocess,
		pluginReadyPreprocessor.Preprocess,
	)
	r.Map("auth:password", injectedHandler(g, &handler.PasswordHandler{}), baseAuthPreprocessors...)

	r.Map("record:fetch", injectedHandler(g, &handler.RecordFetchHandler{}), baseAuthPreprocessors...)
	r.Map("record:query", injectedHandler(g, &handler.RecordQueryHandler{}), baseAuthPreprocessors...)
	r.Map("record:save", injectedHandler(g, &handler.RecordSaveHandler{}), recordWritePreprocessors...)
	r.Map("record:delete", injectedHandler(g, &handler.RecordDeleteHandler{}), recordWritePreprocessors...)

	r.Map("device:register", injectedHandler(g, &handler.DeviceRegisterHandler{}), requireUserWritePreprocessors...)

	// subscription shares the same set of preprocessor as record read at the moment
	r.Map("subscription:fetch_all", injectedHandler(g, &handler.SubscriptionFetchAllHandler{}), requireUserWritePreprocessors...)
	r.Map("subscription:delete", injectedHandler(g, &handler.SubscriptionDeleteHandler{}), requireUserWritePreprocessors...)

	// relation shares the same setof preprocessor
	r.Map("relation:query", injectedHandler(g, &handler.RelationQueryHandler{}), baseAuthPreprocessors...)
	r.Map("relation:add", injectedHandler(g, &handler.RelationAddHandler{}), baseAuthPreprocessors...)
	r.Map("relation:remove", injectedHandler(g, &handler.RelationRemoveHandler{}), baseAuthPreprocessors...)

	r.Map("user:query", injectedHandler(g, &handler.UserQueryHandler{}), baseAuthPreprocessors...)
	r.Map("user:update", injectedHandler(g, &handler.UserUpdateHandler{}), requireUserWritePreprocessors...)
	r.Map("user:link", injectedHandler(g, &handler.UserLinkHandler{}), requireUserWritePreprocessors...)

	r.Map("push:user", injectedHandler(g, &handler.PushToUserHandler{}), notificationPreprocessors...)
	r.Map("push:device", injectedHandler(g, &handler.PushToDeviceHandler{}), notificationPreprocessors...)

	// Following section is for Gateway
	pubSub := pubsub.NewWsPubsub(nil)
	pubSubGateway := router.NewGateway(`pubSub`)
	pubSubGateway.GET(&handler.PubSubHandler{pubSub}, pubSubPreprocessors...)

	internalPubSub := pubsub.NewWsPubsub(internalHub)
	internalPubSubGateway := router.NewGateway(`internalpubSub`)
	internalPubSubGateway.GET(&handler.PubSubHandler{internalPubSub}, pubSubPreprocessors...)

	http.Handle("/", router.LoggingMiddleware(r, false))
	http.Handle("/pubsub", router.LoggingMiddleware(pubSubGateway, false))
	http.Handle("/_/pubsub", router.LoggingMiddleware(internalPubSubGateway, false))

	fileGateway := router.NewGateway(`files/(.+)`)
	fileGateway.GET(injectedHandler(g, &handler.AssetGetURLHandler{}), assetGetPreprocessors...)
	fileGateway.PUT(injectedHandler(g, &handler.AssetUploadURLHandler{}), assetPutPreprocessors...)
	http.Handle("/files/", router.LoggingMiddleware(fileGateway, true))

	// Bootstrap finished, binding port.
	log.Printf("Listening on %v...", config.HTTP.Host)
	err := http.ListenAndServe(config.HTTP.Host, nil)
	if err != nil {
		log.Printf("Failed: %v", err)
	}
}

func injectedHandler(g *inject.Graph, h router.Handler) router.Handler {
	err := g.Provide(&inject.Object{Value: h})
	if err != nil {
		panic(fmt.Sprintf("Unable to set up handler: %v", err))
	}

	err = g.Populate()
	if err != nil {
		panic(fmt.Sprintf("Unable to set up handler: %v", err))
	}
	return h
}

func ensureDB(config Configuration) func() (skydb.Conn, error) {
	connOpener := func() (skydb.Conn, error) {
		return skydb.Open(
			config.DB.ImplName,
			config.App.Name,
			config.App.AccessControl,
			config.DB.Option,
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

func initAssetStore(config Configuration) asset.Store {
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

func initDevice(config Configuration, connOpener func() (skydb.Conn, error)) {
	// TODO: Create a device service to check APNs to remove obsolete devices.
	// The current implementaion deletes pubsub devices if the last registered
	// time is more than 1 day old.
	conn, err := connOpener()
	if err != nil {
		log.Warnf("Failed to delete outdated devices: %v", err)
	}

	conn.DeleteEmptyDevicesByTime(time.Now().AddDate(0, 0, -1))
}

func initPushSender(config Configuration, connOpener func() (skydb.Conn, error)) push.Sender {
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

func initAPNSPusher(config Configuration, connOpener func() (skydb.Conn, error)) *push.APNSPusher {
	apnsPushSender, err := push.NewAPNSPusher(connOpener, push.GatewayType(config.APNS.Env), config.APNS.Cert, config.APNS.Key)
	if err != nil {
		log.Fatalf("Failed to set up push sender: %v", err)
	}
	go apnsPushSender.Run()
	go apnsPushSender.RunFeedback()

	return apnsPushSender
}

func initGCMPusher(config Configuration) *push.GCMPusher {
	return &push.GCMPusher{APIKey: config.GCM.APIKey}
}

func initSubscription(config Configuration, connOpener func() (skydb.Conn, error), hub *pubsub.Hub, pushSender push.Sender) {
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

func initPlugin(config Configuration, initContext *plugin.InitContext) {
	for _, pluginConfig := range config.Plugin {
		initContext.AddPluginConfiguration(pluginConfig.Transport, pluginConfig.Path, pluginConfig.Args)
	}

	initContext.InitPlugins()
}

func initLogger(config Configuration) {
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

func initSentry(config Configuration) {
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
