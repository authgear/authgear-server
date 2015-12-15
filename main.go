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

	connOpener := func() (skydb.Conn, error) {
		return skydb.Open(config.DB.ImplName, config.App.Name, config.DB.Option)
	}
	conn, connError := connOpener()
	if connError != nil {
		log.Fatalf("Failed to start skygear: %v", connError)
	}
	conn.Close()

	pushSender := initPushSender(config, connOpener)

	internalHub := pubsub.NewHub()
	initSubscription(config, connOpener, internalHub, pushSender)
	initDevice(config, connOpener)

	notificationPreprocessor := notificationPreprocessor{
		NotificationSender: pushSender,
	}

	store := initAssetStore(config)
	assetStorePreprocessor := assetStorePreprocessor{
		Store: store,
	}

	naiveAPIKeyPreprocessor := apiKeyValidatonPreprocessor{
		Key:     config.App.APIKey,
		AppName: config.App.Name,
	}

	tokenStore := authtoken.InitTokenStore(config.TokenStore.ImplName, config.TokenStore.Path)
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

	assetGetPreprocessors := []router.Processor{
		dbConnPreprocessor.Preprocess,
		assetStorePreprocessor.Preprocess,
	}
	assetUploadPreprocessors := []router.Processor{
		authenticator.Preprocess,
		dbConnPreprocessor.Preprocess,
		assetStorePreprocessor.Preprocess,
	}
	fileGateway := router.NewGateway(`files/(.+)`)
	fileGateway.GET(handler.AssetGetURLHandler, assetGetPreprocessors...)
	fileGateway.PUT(handler.AssetUploadURLHandler, assetUploadPreprocessors...)
	http.Handle("/files/", router.LoggingMiddleware(fileGateway, true))

	r := router.NewRouter()
	r.Map("", handler.HomeHandler)

	providerRegistry := provider.NewRegistry()
	providerRegistryPreprocessor := providerRegistryPreprocessor{
		Registry: providerRegistry,
	}

	authPreprocessors := []router.Processor{
		naiveAPIKeyPreprocessor.Preprocess,
		dbConnPreprocessor.Preprocess,
		tokenStorePreprocessor.Preprocess,
		providerRegistryPreprocessor.Preprocess,
	}
	r.Map("auth:signup", handler.SignupHandler, authPreprocessors...)
	r.Map("auth:login", handler.LoginHandler, authPreprocessors...)
	r.Map("auth:logout", handler.LogoutHandler,
		authenticator.Preprocess,
		providerRegistryPreprocessor.Preprocess,
	)
	r.Map("auth:password", handler.PasswordHandler,
		dbConnPreprocessor.Preprocess,
		authenticator.Preprocess,
	)

	hookRegistry := hook.NewRegistry()
	hookRegistryPreprocessor := hookRegistryPreprocessor{
		Registry: hookRegistry,
	}

	recordReadPreprocessors := []router.Processor{
		authenticator.Preprocess,
		dbConnPreprocessor.Preprocess,
		assetStorePreprocessor.Preprocess,
		injectUserIfPresent,
		injectDatabase,
	}
	recordWritePreprocessors := []router.Processor{
		hookRegistryPreprocessor.Preprocess,
		authenticator.Preprocess,
		dbConnPreprocessor.Preprocess,
		assetStorePreprocessor.Preprocess,
		injectUserIfPresent,
		injectDatabase,
		requireUserForWrite,
	}
	r.Map("record:fetch", handler.RecordFetchHandler, recordReadPreprocessors...)
	r.Map("record:query", handler.RecordQueryHandler, recordReadPreprocessors...)
	r.Map("record:save", handler.RecordSaveHandler, recordWritePreprocessors...)
	r.Map("record:delete", handler.RecordDeleteHandler, recordWritePreprocessors...)

	r.Map("device:register",
		handler.DeviceRegisterHandler,
		authenticator.Preprocess,
		dbConnPreprocessor.Preprocess,
		injectUserIfPresent,
	)

	// subscription shares the same set of preprocessor as record read at the moment
	r.Map("subscription:fetch_all", handler.SubscriptionFetchAllHandler, recordReadPreprocessors...)
	r.Map("subscription:fetch", handler.SubscriptionFetchHandler, recordReadPreprocessors...)
	r.Map("subscription:save", handler.SubscriptionSaveHandler, recordReadPreprocessors...)
	r.Map("subscription:delete", handler.SubscriptionDeleteHandler, recordReadPreprocessors...)

	// relation shares the same setof preprocessor
	r.Map("relation:query", handler.RelationQueryHandler, recordReadPreprocessors...)
	r.Map("relation:add", handler.RelationAddHandler, recordReadPreprocessors...)
	r.Map("relation:remove", handler.RelationRemoveHandler, recordReadPreprocessors...)

	userReadPreprocessors := []router.Processor{
		authenticator.Preprocess,
		dbConnPreprocessor.Preprocess,
		injectUserIfPresent,
		injectDatabase,
	}
	userWritePreprocessors := []router.Processor{
		authenticator.Preprocess,
		dbConnPreprocessor.Preprocess,
		injectUserIfPresent,
		injectDatabase,
		requireUserForWrite,
	}
	r.Map("user:query", handler.UserQueryHandler, userReadPreprocessors...)
	r.Map("user:update", handler.UserUpdateHandler, userWritePreprocessors...)
	r.Map("user:link", handler.UserLinkHandler,
		authenticator.Preprocess,
		dbConnPreprocessor.Preprocess,
		providerRegistryPreprocessor.Preprocess,
		injectUserIfPresent,
		injectDatabase,
		requireUserForWrite,
	)

	notificationPreprocessors := []router.Processor{
		naiveAPIKeyPreprocessor.Preprocess,
		dbConnPreprocessor.Preprocess,
		injectDatabase,
		notificationPreprocessor.Preprocess,
	}

	r.Map("push:user", handler.PushToUserHandler, notificationPreprocessors...)
	r.Map("push:device", handler.PushToDeviceHandler, notificationPreprocessors...)

	plugins := []plugin.Plugin{}
	for _, pluginConfig := range config.Plugin {
		p := plugin.NewPlugin(pluginConfig.Transport, pluginConfig.Path, pluginConfig.Args)

		plugins = append(plugins, p)
	}

	c := cron.New()
	initContext := plugin.InitContext{
		Router:           r,
		HookRegistry:     hookRegistry,
		ProviderRegistry: providerRegistry,
		Scheduler:        c,
	}

	for _, plug := range plugins {
		plug.Init(&initContext)
	}
	c.Start()

	pubSubPreprocessors := []router.Processor{
		naiveAPIKeyPreprocessor.Preprocess,
	}

	pubSub := pubsub.NewWsPubsub(nil)
	pubSubGateway := router.NewGateway(`pubSub`)
	pubSubGateway.GET(handler.NewPubSubHandler(pubSub), pubSubPreprocessors...)

	internalPubSub := pubsub.NewWsPubsub(internalHub)
	internalPubSubGateway := router.NewGateway(`internalpubSub`)
	internalPubSubGateway.GET(handler.NewPubSubHandler(internalPubSub), pubSubPreprocessors...)

	http.Handle("/", router.LoggingMiddleware(r, false))
	http.Handle("/pubsub", router.LoggingMiddleware(pubSubGateway, false))
	http.Handle("/_/pubsub", router.LoggingMiddleware(internalPubSubGateway, false))

	log.Printf("Listening on %v...", config.HTTP.Host)
	err := http.ListenAndServe(config.HTTP.Host, nil)
	if err != nil {
		log.Printf("Failed: %v", err)
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
			config.AssetURLSigner.Secret)
	case "s3":
		s3Store, err := asset.NewS3Store(
			config.AssetStore.AccessToken,
			config.AssetStore.SecretToken,
			config.AssetStore.Reigon,
			config.AssetStore.Bucket,
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
		routeSender.Route("aps", initAPNSPusher(config, connOpener))
	}
	if config.GCM.Enable {
		routeSender.Route("gcm", initGCMPusher(config))
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
