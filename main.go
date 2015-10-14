package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/Sirupsen/logrus/hooks/sentry"
	"github.com/robfig/cron"

	"github.com/oursky/ourd/asset"
	"github.com/oursky/ourd/authtoken"
	"github.com/oursky/ourd/handler"
	"github.com/oursky/ourd/hook"
	"github.com/oursky/ourd/oddb"
	_ "github.com/oursky/ourd/oddb/fs"
	_ "github.com/oursky/ourd/oddb/pq"
	"github.com/oursky/ourd/plugin"
	_ "github.com/oursky/ourd/plugin/exec"
	_ "github.com/oursky/ourd/plugin/zmq"
	"github.com/oursky/ourd/provider"
	"github.com/oursky/ourd/pubsub"
	"github.com/oursky/ourd/push"
	"github.com/oursky/ourd/router"
	"github.com/oursky/ourd/subscription"
)

type responseLogger struct {
	w      http.ResponseWriter
	status int
	size   int
	b      bytes.Buffer
}

func (l *responseLogger) Header() http.Header {
	return l.w.Header()
}

func (l *responseLogger) Write(b []byte) (int, error) {
	if l.status == 0 {
		// The status will be StatusOK if WriteHeader has not been called yet
		l.status = http.StatusOK
	}
	l.b.Write(b)
	size, err := l.w.Write(b)
	l.size += size
	return size, err
}

func (l *responseLogger) WriteHeader(s int) {
	l.w.WriteHeader(s)
	l.status = s
}

func (l *responseLogger) Status() int {
	return l.status
}

func (l *responseLogger) Size() int {
	return l.size
}

func (l *responseLogger) String() string {
	return l.b.String()
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("%v %v", r.Method, r.RequestURI)

		log.Debugln("------ Header: ------")
		for key, value := range r.Header {
			log.Debugf("%s: %v", key, value)
		}

		body, _ := ioutil.ReadAll(r.Body)
		r.Body = ioutil.NopCloser(bytes.NewReader(body))

		log.Debugln("------ Request: ------")
		if r.Header.Get("Content-Type") == "" || r.Header.Get("Content-Type") == "application/json" {
			log.Debugln(string(body))
		} else {
			log.Debugf("%d bytes of body", len(body))
		}

		rlogger := &responseLogger{w: w}
		next.ServeHTTP(rlogger, r)

		log.Debugln("------ Response: ------")
		if w.Header().Get("Content-Type") == "" || w.Header().Get("Content-Type") == "application/json" {
			log.Debugln(rlogger.String())
		} else {
			log.Debugf("%d bytes of body", len(rlogger.String()))
		}
	})
}

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

	initLogger(config)

	connOpener := func() (oddb.Conn, error) { return oddb.Open(config.DB.ImplName, config.App.Name, config.DB.Option) }

	var pushSender push.Sender
	if config.APNS.Enable {
		apnsPushSender, err := push.NewAPNSPusher(connOpener, push.GatewayType(config.APNS.Env), config.APNS.Cert, config.APNS.Key)
		if err != nil {
			log.Fatalf("Failed to set up push sender: %v", err)
		}
		go apnsPushSender.Run()
		go apnsPushSender.RunFeedback()
		pushSender = apnsPushSender
	}

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

	fileTokenStorePreprocessor := tokenStorePreprocessor{
		Store: authtoken.FileStore(config.TokenStore.Path).Init(),
	}

	authenticator := userAuthenticator{
		APIKey:  config.App.APIKey,
		AppName: config.App.Name,
	}

	fileSystemConnPreprocessor := connPreprocessor{
		AppName:  config.App.Name,
		DBOpener: oddb.Open,
		DBImpl:   config.DB.ImplName,
		Option:   config.DB.Option,
	}

	assetGetPreprocessors := []router.Processor{
		fileSystemConnPreprocessor.Preprocess,
		assetStorePreprocessor.Preprocess,
	}
	assetUploadPreprocessors := []router.Processor{
		fileTokenStorePreprocessor.Preprocess,
		authenticator.Preprocess,
		fileSystemConnPreprocessor.Preprocess,
		assetStorePreprocessor.Preprocess,
	}
	fileGateway := router.NewGateway(`files/(.+)`)
	fileGateway.GET(handler.AssetGetURLHandler, assetGetPreprocessors...)
	fileGateway.PUT(handler.AssetUploadURLHandler, assetUploadPreprocessors...)
	http.Handle("/files/", logMiddleware(fileGateway))

	r := router.NewRouter()
	r.Map("", handler.HomeHandler)

	providerRegistry := provider.NewRegistry()
	providerRegistryPreprocessor := providerRegistryPreprocessor{
		Registry: providerRegistry,
	}

	authPreprocessors := []router.Processor{
		naiveAPIKeyPreprocessor.Preprocess,
		fileSystemConnPreprocessor.Preprocess,
		fileTokenStorePreprocessor.Preprocess,
		providerRegistryPreprocessor.Preprocess,
	}
	r.Map("auth:signup", handler.SignupHandler, authPreprocessors...)
	r.Map("auth:login", handler.LoginHandler, authPreprocessors...)
	r.Map("auth:logout", handler.LogoutHandler,
		fileTokenStorePreprocessor.Preprocess,
		authenticator.Preprocess,
		providerRegistryPreprocessor.Preprocess,
	)
	r.Map("auth:password", handler.PasswordHandler,
		fileSystemConnPreprocessor.Preprocess,
		fileTokenStorePreprocessor.Preprocess,
		authenticator.Preprocess,
	)

	hookRegistry := hook.NewRegistry()
	hookRegistryPreprocessor := hookRegistryPreprocessor{
		Registry: hookRegistry,
	}

	recordReadPreprocessors := []router.Processor{
		fileTokenStorePreprocessor.Preprocess,
		authenticator.Preprocess,
		fileSystemConnPreprocessor.Preprocess,
		assetStorePreprocessor.Preprocess,
		injectUserIfPresent,
		injectDatabase,
	}
	recordWritePreprocessors := []router.Processor{
		hookRegistryPreprocessor.Preprocess,
		fileTokenStorePreprocessor.Preprocess,
		authenticator.Preprocess,
		fileSystemConnPreprocessor.Preprocess,
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
		fileTokenStorePreprocessor.Preprocess,
		authenticator.Preprocess,
		fileSystemConnPreprocessor.Preprocess,
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
		fileTokenStorePreprocessor.Preprocess,
		authenticator.Preprocess,
		fileSystemConnPreprocessor.Preprocess,
		injectUserIfPresent,
		injectDatabase,
	}
	userWritePreprocessors := []router.Processor{
		fileTokenStorePreprocessor.Preprocess,
		authenticator.Preprocess,
		fileSystemConnPreprocessor.Preprocess,
		injectUserIfPresent,
		injectDatabase,
		requireUserForWrite,
	}
	r.Map("user:query", handler.UserQueryHandler, userReadPreprocessors...)
	r.Map("user:update", handler.UserUpdateHandler, userWritePreprocessors...)
	r.Map("user:link", handler.UserLinkHandler,
		fileTokenStorePreprocessor.Preprocess,
		authenticator.Preprocess,
		fileSystemConnPreprocessor.Preprocess,
		providerRegistryPreprocessor.Preprocess,
		injectUserIfPresent,
		injectDatabase,
		requireUserForWrite,
	)

	notificationPreprocessors := []router.Processor{
		naiveAPIKeyPreprocessor.Preprocess,
		fileSystemConnPreprocessor.Preprocess,
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

	http.Handle("/", logMiddleware(r))
	http.Handle("/pubsub", pubSubGateway)
	http.Handle("/_/pubsub", internalPubSubGateway)

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

func initDevice(config Configuration, connOpener func() (oddb.Conn, error)) {
	// TODO: Create a device service to check APNs to remove obsolete devices.
	// The current implementaion deletes pubsub devices if the last registered
	// time is more than 1 day old.
	conn, err := connOpener()
	if err != nil {
		log.Warnf("Failed to delete outdated devices: %v", err)
	}

	conn.DeleteDeviceByType("pubsub", time.Now().AddDate(0, 0, -1))
}

func initSubscription(config Configuration, connOpener func() (oddb.Conn, error), hub *pubsub.Hub, pushSender push.Sender) {
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
