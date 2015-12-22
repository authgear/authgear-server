package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/Sirupsen/logrus/hooks/sentry"
	"github.com/robfig/cron"

	"github.com/oursky/skygear/asset"
	"github.com/oursky/skygear/authtoken"
	"github.com/oursky/skygear/handler"
	"github.com/oursky/skygear/hook"
	"github.com/oursky/skygear/plugin"
	_ "github.com/oursky/skygear/plugin/exec"
	_ "github.com/oursky/skygear/plugin/zmq"
	"github.com/oursky/skygear/provider"
	"github.com/oursky/skygear/pubsub"
	"github.com/oursky/skygear/push"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skydb"
	_ "github.com/oursky/skygear/skydb/fs"
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
	store := initAssetStore(config)
	tokenStore := authtoken.InitTokenStore(config.TokenStore.ImplName, config.TokenStore.Path)

	providerRegistry := provider.NewRegistry()
	hookRegistry := hook.NewRegistry()
	c := cron.New()
	initContext := plugin.InitContext{
		Router:           r,
		HookRegistry:     hookRegistry,
		ProviderRegistry: providerRegistry,
		Scheduler:        c,
	}
	c.Start()
	initPlugin(config, &initContext) // Block until plugin configured

	internalHub := pubsub.NewHub()
	initSubscription(config, connOpener, internalHub, pushSender)
	initDevice(config, connOpener)

	// Preprocessor
	notificationPreprocessor := notificationPreprocessor{
		NotificationSender: pushSender,
	}

	assetStorePreprocessor := assetStorePreprocessor{
		Store: store,
	}

	naiveAPIKeyPreprocessor := apiKeyValidatonPreprocessor{
		Key:     config.App.APIKey,
		AppName: config.App.Name,
	}

	tokenStorePreprocessor := tokenStorePreprocessor{
		Store: tokenStore,
	}

	authenticator := userAuthenticator{
		APIKey:     config.App.APIKey,
		AppName:    config.App.Name,
		TokenStore: tokenStore,
	}

	dbConnPreprocessor := connPreprocessor{
		AppName:  config.App.Name,
		DBOpener: skydb.Open,
		DBImpl:   config.DB.ImplName,
		Option:   config.DB.Option,
	}

	providerRegistryPreprocessor := providerRegistryPreprocessor{
		Registry: providerRegistry,
	}

	hookRegistryPreprocessor := hookRegistryPreprocessor{
		Registry: hookRegistry,
	}

	baseAuthPreprocessors := []router.Processor{
		authenticator.Preprocess,
		dbConnPreprocessor.Preprocess,
		assetStorePreprocessor.Preprocess,
		injectUserIfPresent,
		injectDatabase,
	}

	assetGetPreprocessors := []router.Processor{
		dbConnPreprocessor.Preprocess,
		assetStorePreprocessor.Preprocess,
	}

	authPreprocessors := []router.Processor{
		naiveAPIKeyPreprocessor.Preprocess,
		dbConnPreprocessor.Preprocess,
		tokenStorePreprocessor.Preprocess,
		providerRegistryPreprocessor.Preprocess,
	}

	recordWritePreprocessors := append(baseAuthPreprocessors,
		hookRegistryPreprocessor.Preprocess,
		requireUserForWrite,
	)

	userWritePreprocessors := append(baseAuthPreprocessors,
		requireUserForWrite,
	)

	userLinkPreprocessors := append(baseAuthPreprocessors,
		providerRegistryPreprocessor.Preprocess,
		requireUserForWrite,
	)

	devicePreprocessors := append(baseAuthPreprocessors,
		requireUserForWrite,
	)

	subscriptionPreprocessors := append(baseAuthPreprocessors,
		requireUserForWrite,
	)

	notificationPreprocessors := []router.Processor{
		naiveAPIKeyPreprocessor.Preprocess,
		dbConnPreprocessor.Preprocess,
		injectDatabase,
		notificationPreprocessor.Preprocess,
	}

	pubSubPreprocessors := []router.Processor{
		naiveAPIKeyPreprocessor.Preprocess,
	}

	r.Map("", handler.HomeHandler)

	r.Map("auth:signup", handler.SignupHandler, authPreprocessors...)
	r.Map("auth:login", handler.LoginHandler, authPreprocessors...)
	r.Map("auth:logout", handler.LogoutHandler,
		authenticator.Preprocess,
		providerRegistryPreprocessor.Preprocess,
	)
	r.Map("auth:password", handler.PasswordHandler, baseAuthPreprocessors...)

	r.Map("record:fetch", handler.RecordFetchHandler, baseAuthPreprocessors...)
	r.Map("record:query", handler.RecordQueryHandler, baseAuthPreprocessors...)
	r.Map("record:save", handler.RecordSaveHandler, recordWritePreprocessors...)
	r.Map("record:delete", handler.RecordDeleteHandler, recordWritePreprocessors...)

	r.Map("device:register", handler.DeviceRegisterHandler, devicePreprocessors...)

	// subscription shares the same set of preprocessor as record read at the moment
	r.Map("subscription:fetch_all", handler.SubscriptionFetchAllHandler, subscriptionPreprocessors...)
	r.Map("subscription:fetch", handler.SubscriptionFetchHandler, subscriptionPreprocessors...)
	r.Map("subscription:save", handler.SubscriptionSaveHandler, subscriptionPreprocessors...)
	r.Map("subscription:delete", handler.SubscriptionDeleteHandler, subscriptionPreprocessors...)

	// relation shares the same setof preprocessor
	r.Map("relation:query", handler.RelationQueryHandler, baseAuthPreprocessors...)
	r.Map("relation:add", handler.RelationAddHandler, baseAuthPreprocessors...)
	r.Map("relation:remove", handler.RelationRemoveHandler, baseAuthPreprocessors...)

	r.Map("user:query", handler.UserQueryHandler, baseAuthPreprocessors...)
	r.Map("user:update", handler.UserUpdateHandler, userWritePreprocessors...)
	r.Map("user:link", handler.UserLinkHandler, userLinkPreprocessors...)

	r.Map("push:user", handler.PushToUserHandler, notificationPreprocessors...)
	r.Map("push:device", handler.PushToDeviceHandler, notificationPreprocessors...)

	// Following section is for Gateway
	pubSub := pubsub.NewWsPubsub(nil)
	pubSubGateway := router.NewGateway(`pubSub`)
	pubSubGateway.GET(handler.NewPubSubHandler(pubSub), pubSubPreprocessors...)

	internalPubSub := pubsub.NewWsPubsub(internalHub)
	internalPubSubGateway := router.NewGateway(`internalpubSub`)
	internalPubSubGateway.GET(handler.NewPubSubHandler(internalPubSub), pubSubPreprocessors...)

	http.Handle("/", router.LoggingMiddleware(r, false))
	http.Handle("/pubsub", router.LoggingMiddleware(pubSubGateway, false))
	http.Handle("/_/pubsub", router.LoggingMiddleware(internalPubSubGateway, false))

	fileGateway := router.NewGateway(`files/(.+)`)
	fileGateway.GET(handler.AssetGetURLHandler, assetGetPreprocessors...)
	fileGateway.PUT(handler.AssetUploadURLHandler, baseAuthPreprocessors...)
	http.Handle("/files/", router.LoggingMiddleware(fileGateway, true))

	// Bootstrap finished, binding port.
	log.Printf("Listening on %v...", config.HTTP.Host)
	err := http.ListenAndServe(config.HTTP.Host, nil)
	if err != nil {
		log.Printf("Failed: %v", err)
	}
}

func ensureDB(config Configuration) func() (skydb.Conn, error) {
	connOpener := func() (skydb.Conn, error) {
		return skydb.Open(config.DB.ImplName, config.App.Name, config.DB.Option)
	}
	conn, connError := connOpener()
	if connError != nil {
		log.Fatalf("Failed to start skygear: %v", connError)
	}
	conn.Close()
	return connOpener
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

	if routeSender.Len() == 0 {
		return nil
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
	plugins := []plugin.Plugin{}
	for _, pluginConfig := range config.Plugin {
		p := plugin.NewPlugin(pluginConfig.Transport, pluginConfig.Path, pluginConfig.Args)

		plugins = append(plugins, p)
	}

	for _, plug := range plugins {
		plug.Init(initContext)
	}
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
