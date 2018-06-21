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
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/evalphobia/logrus_sentry"
	"github.com/facebookgo/inject"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/server/asset"
	"github.com/skygeario/skygear-server/pkg/server/audit"
	"github.com/skygeario/skygear-server/pkg/server/authtoken"
	"github.com/skygeario/skygear-server/pkg/server/handler"
	"github.com/skygeario/skygear-server/pkg/server/logging"
	"github.com/skygeario/skygear-server/pkg/server/plugin"
	pluginEvent "github.com/skygeario/skygear-server/pkg/server/plugin/event"
	_ "github.com/skygeario/skygear-server/pkg/server/plugin/exec"
	"github.com/skygeario/skygear-server/pkg/server/plugin/hook"
	_ "github.com/skygeario/skygear-server/pkg/server/plugin/http"
	"github.com/skygeario/skygear-server/pkg/server/plugin/provider"
	_ "github.com/skygeario/skygear-server/pkg/server/plugin/zmq"
	pp "github.com/skygeario/skygear-server/pkg/server/preprocessor"
	"github.com/skygeario/skygear-server/pkg/server/pubsub"
	"github.com/skygeario/skygear-server/pkg/server/push"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skyconfig"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	_ "github.com/skygeario/skygear-server/pkg/server/skydb/pq"
	"github.com/skygeario/skygear-server/pkg/server/skyversion"
	"github.com/skygeario/skygear-server/pkg/server/subscription"
)

var log = logging.LoggerEntry("main")

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "version" {
			fmt.Printf("%s\n", skyversion.Version())
			os.Exit(0)
		}
	}

	config := skyconfig.NewConfiguration()
	config.ReadFromEnv()
	if err := config.Validate(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	initLogger(config)

	mainLogger := logging.LoggerEntryWithTag("main", "") // untagged logger
	mainLogger.Infof("Starting Skygear Server(%s)...", skyversion.Version())
	connOpener := ensureDB(config) // Fatal on DB failed

	initUserAuthRecordKeys(connOpener, config.App.AuthRecordKeys)

	if config.App.Slave {
		mainLogger.Infof("Skygear Server is running in slave mode.")
	}

	// Init all the services
	r := router.NewRouter()
	r.ResponseTimeout = time.Duration(config.App.ResponseTimeout) * time.Second
	serveMux := http.NewServeMux()
	pushSender := initPushSender(config, connOpener)

	tokenStore := authtoken.InitTokenStore(authtoken.Configuration{
		Implementation: config.TokenStore.ImplName,
		Path:           config.TokenStore.Path,
		Prefix:         config.TokenStore.Prefix,
		Expiry:         config.TokenStore.Expiry,
		Secret:         config.TokenStore.Secret,
	})

	dbConfig := baseDBConfig(config)

	passwordChecker := &audit.PasswordChecker{
		PwMinLength:            config.UserAudit.PwMinLength,
		PwUppercaseRequired:    config.UserAudit.PwUppercaseRequired,
		PwLowercaseRequired:    config.UserAudit.PwLowercaseRequired,
		PwDigitRequired:        config.UserAudit.PwDigitRequired,
		PwSymbolRequired:       config.UserAudit.PwSymbolRequired,
		PwMinGuessableLevel:    config.UserAudit.PwMinGuessableLevel,
		PwExcludedKeywords:     config.UserAudit.PwExcludedKeywords,
		PwExcludedFields:       config.UserAudit.PwExcludedFields,
		PwHistorySize:          config.UserAudit.PwHistorySize,
		PwHistoryDays:          config.UserAudit.PwHistoryDays,
		PasswordHistoryEnabled: dbConfig.PasswordHistoryEnabled,
	}

	pwHousekeeper := &audit.PwHousekeeper{
		AppName:       config.App.Name,
		AccessControl: config.App.AccessControl,
		DBOpener:      skydb.Open,
		DBImpl:        config.DB.ImplName,
		Option:        config.DB.Option,
		DBConfig:      dbConfig,

		PwHistorySize:          config.UserAudit.PwHistorySize,
		PwHistoryDays:          config.UserAudit.PwHistoryDays,
		PasswordHistoryEnabled: dbConfig.PasswordHistoryEnabled,
	}

	preprocessorRegistry := router.PreprocessorRegistry{}

	var cronjob *cron.Cron
	if !config.App.Slave {
		cronjob = cron.New()
	}
	pluginContext := plugin.Context{
		Router:           r,
		Mux:              serveMux,
		HookRegistry:     hook.NewRegistry(),
		ProviderRegistry: provider.NewRegistry(),
		Scheduler:        cronjob,
		Config:           config,
	}

	var internalHub *pubsub.Hub
	if !config.App.Slave {
		internalHub = pubsub.NewHub()
		initSubscription(config, connOpener, internalHub, pushSender)
		initDevice(config, connOpener)
	}

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
		ClientKey:          config.App.APIKey,
		MasterKey:          config.App.MasterKey,
		AppName:            config.App.Name,
		TokenStore:         tokenStore,
		BypassUnauthorized: false,
	}
	preprocessorRegistry["inject_auth_id"] = &pp.UserAuthenticator{
		ClientKey:          config.App.APIKey,
		MasterKey:          config.App.MasterKey,
		AppName:            config.App.Name,
		TokenStore:         tokenStore,
		BypassUnauthorized: true,
	}
	preprocessorRegistry["dbconn"] = &pp.ConnPreprocessor{
		AppName:       config.App.Name,
		AccessControl: config.App.AccessControl,
		DBOpener:      skydb.Open,
		DBImpl:        config.DB.ImplName,
		Option:        config.DB.Option,
		DBConfig:      dbConfig,
	}
	preprocessorRegistry["plugin_ready"] = &pp.EnsurePluginReadyPreprocessor{
		PluginContext: &pluginContext,
		ClientKey:     config.App.APIKey,
		MasterKey:     config.App.MasterKey,
	}
	preprocessorRegistry["inject_auth"] = &pp.InjectAuth{
		PwExpiryDays: config.UserAudit.PwExpiryDays,
	}
	preprocessorRegistry["require_auth"] = &pp.InjectAuth{
		PwExpiryDays: config.UserAudit.PwExpiryDays,
		Required:     true,
	}
	preprocessorRegistry["require_auth_skip_pwexpiry"] = &pp.InjectAuth{
		PwExpiryDays: 0, // skipping password expiry check
		Required:     true,
	}
	preprocessorRegistry["require_user"] = &pp.InjectUser{
		Required:          true,
		CheckVerification: config.Verification.Required,
	}
	if config.Verification.Required {
		preprocessorRegistry["check_user"] = &pp.InjectUser{
			Required:          false,
			CheckVerification: true,
		}
	} else {
		preprocessorRegistry["check_user"] = &pp.Null{}
	}
	preprocessorRegistry["require_admin"] = &pp.RequireAdminOrMasterKey{}
	preprocessorRegistry["require_master_key"] = &pp.RequireMasterKey{}
	preprocessorRegistry["inject_db"] = &pp.InjectDatabase{}
	preprocessorRegistry["inject_public_db"] = &pp.InjectPublicDatabase{}
	preprocessorRegistry["dev_only"] = &pp.DevOnlyProcessor{
		DevMode: config.App.DevMode,
	}

	g := &inject.Graph{}
	injectErr := g.Provide(
		&inject.Object{
			Value:    pluginContext.ProviderRegistry,
			Complete: true,
			Name:     "ProviderRegistry",
		},
		&inject.Object{
			Value:    pluginContext.HookRegistry,
			Complete: true,
			Name:     "HookRegistry",
		},
		&inject.Object{
			Value:    tokenStore,
			Complete: true,
			Name:     "TokenStore",
		},
		&inject.Object{
			Value:    initAssetStore(config),
			Complete: true,
			Name:     "AssetStore",
		},
		&inject.Object{
			Value:    pushSender,
			Complete: true,
			Name:     "PushSender",
		},
		&inject.Object{
			Value:    pluginEvent.NewSender(&pluginContext),
			Complete: true,
			Name:     "PluginEventSender",
		},
		&inject.Object{
			Value:    skydb.GetAccessModel(config.App.AccessControl),
			Complete: true,
			Name:     "AccessModel",
		},
		&inject.Object{
			Value:    config.App.AuthRecordKeys,
			Complete: true,
			Name:     "AuthRecordKeys",
		},
		&inject.Object{
			Value:    passwordChecker,
			Complete: true,
			Name:     "PasswordChecker",
		},
		&inject.Object{
			Value:    pwHousekeeper,
			Complete: true,
			Name:     "PwHousekeeper",
		},
	)
	if injectErr != nil {
		panic(fmt.Sprintf("Unable to set up handler: %v", injectErr))
	}

	injector := router.HandlerInjector{
		ServiceGraph:    g,
		PreprocessorMap: &preprocessorRegistry,
	}
	pluginContext.HandlerInjector = injector

	r.Map("", "", &handler.HomeHandler{})
	r.Map("_status:healthz", "", injector.Inject(&handler.HealthzHandler{}))

	r.Map("auth:signup", "auth", injector.Inject(&handler.SignupHandler{}))
	r.Map("auth:login", "auth", injector.Inject(&handler.LoginHandler{}))
	r.Map("auth:logout", "auth", injector.Inject(&handler.LogoutHandler{}))
	r.Map("auth:password", "auth", injector.Inject(&handler.ChangePasswordHandler{}))
	r.Map("auth:reset_password", "auth", injector.Inject(&handler.ResetPasswordHandler{}))
	r.Map("auth:disable:set", "auth", injector.Inject(&handler.SetDisableUserHandler{}))
	r.Map("sso:oauth:login", "sso", injector.Inject(&handler.LoginProviderHandler{}))
	r.Map("sso:oauth:signup", "sso", injector.Inject(&handler.SignupProviderHandler{}))
	r.Map("sso:oauth:link", "sso", injector.Inject(&handler.LinkProviderHandler{}))
	r.Map("sso:oauth:unlink", "sso", injector.Inject(&handler.UnlinkProviderHandler{}))
	r.Map("sso:custom_token:login", "sso", injector.Inject(&handler.SSOCustomTokenLoginHandler{
		CustomTokenSecret: config.Auth.CustomTokenSecret,
	}))

	r.Map("asset:put", "asset", injector.Inject(&handler.AssetUploadHandler{}))

	r.Map("record:fetch", "record", injector.Inject(&handler.RecordFetchHandler{}))
	r.Map("record:query", "record", injector.Inject(&handler.RecordQueryHandler{}))
	r.Map("record:save", "record", injector.Inject(&handler.RecordSaveHandler{}))
	r.Map("record:delete", "record", injector.Inject(&handler.RecordDeleteHandler{}))

	r.Map("device:register", "device", injector.Inject(&handler.DeviceRegisterHandler{}))
	r.Map("device:unregister", "device", injector.Inject(&handler.DeviceUnregisterHandler{}))

	// subscription shares the same set of preprocessor as record read at the moment
	r.Map("subscription:fetch_all", "subscription", injector.Inject(&handler.SubscriptionFetchAllHandler{}))
	r.Map("subscription:fetch", "subscription", injector.Inject(&handler.SubscriptionFetchHandler{}))
	r.Map("subscription:save", "subscription", injector.Inject(&handler.SubscriptionSaveHandler{}))
	r.Map("subscription:delete", "subscription", injector.Inject(&handler.SubscriptionDeleteHandler{}))

	// relation shares the same setof preprocessor
	r.Map("relation:query", "relation", injector.Inject(&handler.RelationQueryHandler{}))
	r.Map("relation:add", "relation", injector.Inject(&handler.RelationAddHandler{}))
	r.Map("relation:remove", "relation", injector.Inject(&handler.RelationRemoveHandler{}))

	r.Map("me", "", injector.Inject(&handler.MeHandler{}))

	r.Map("role:default", "role", injector.Inject(&handler.RoleDefaultHandler{}))
	r.Map("role:admin", "role", injector.Inject(&handler.RoleAdminHandler{}))
	r.Map("role:assign", "role", injector.Inject(&handler.RoleAssignHandler{}))
	r.Map("role:revoke", "role", injector.Inject(&handler.RoleRevokeHandler{}))
	r.Map("role:get", "role", injector.Inject(&handler.RoleGetHandler{}))

	r.Map("push:user", "push", injector.Inject(&handler.PushToUserHandler{}))
	r.Map("push:device", "push", injector.Inject(&handler.PushToDeviceHandler{}))

	r.Map("schema:rename", "schema", injector.Inject(&handler.SchemaRenameHandler{}))
	r.Map("schema:delete", "schema", injector.Inject(&handler.SchemaDeleteHandler{}))
	r.Map("schema:create", "schema", injector.Inject(&handler.SchemaCreateHandler{}))
	r.Map("schema:fetch", "schema", injector.Inject(&handler.SchemaFetchHandler{}))
	r.Map("schema:access", "schema", injector.Inject(&handler.SchemaAccessHandler{}))
	r.Map("schema:default_access", "schema", injector.Inject(&handler.SchemaDefaultAccessHandler{}))
	r.Map("schema:field_access:get", "schema", injector.Inject(&handler.SchemaFieldAccessGetHandler{}))
	r.Map("schema:field_access:update", "schema", injector.Inject(&handler.SchemaFieldAccessUpdateHandler{}))

	serveMux.Handle("/", r)

	// Following section is for Gateway
	if !config.App.Slave {
		pubSub := pubsub.NewWsPubsub(nil)
		pubSubGateway := router.NewGateway("", "/pubsub", "pubsub", serveMux)
		pubSubGateway.GET(injector.InjectProcessors(&handler.PubSubHandler{
			WebSocket: pubSub,
		}))

		internalPubSub := pubsub.NewWsPubsub(internalHub)
		internalPubSubGateway := router.NewGateway("", "/_/pubsub", "pubsub", serveMux)
		internalPubSubGateway.GET(injector.InjectProcessors(&handler.PubSubHandler{
			WebSocket: internalPubSub,
		}))
	}

	fileGateway := router.NewGateway("files/(.+)", "/files/", "asset", serveMux)
	fileGateway.ResponseTimeout = time.Duration(config.App.ResponseTimeout) * time.Second
	fileGateway.GET(injector.Inject(&handler.GetFileHandler{}))

	uploadFileHandler := injector.Inject(&handler.UploadFileHandler{})
	fileGateway.PUT(uploadFileHandler)
	fileGateway.POST(uploadFileHandler)

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
		loggingMiddleware := &router.LoggingMiddleware{
			Skips: []string{
				"/files/",
				"/_/pubsub/",
				"/pubsub/",
			},
			MimeConcern: []string{
				"",
				"application/json",
			},
			Next: finalMux,
		}

		if config.LOG.RouterByteLimit > 0 {
			var limit int
			limit = int(config.LOG.RouterByteLimit)
			loggingMiddleware.ByteLimit = &limit
		}

		finalMux = loggingMiddleware
	}

	finalMux = &router.RequestIDMiddleware{
		Next: finalMux,
	}

	// Bootstrap finished, starting services
	initPlugin(config, &pluginContext)

	mainLogger.Printf("Listening on %v...", config.HTTP.Host)
	err := http.ListenAndServe(config.HTTP.Host, finalMux)
	if err != nil {
		mainLogger.Printf("Failed: %v", err)
		os.Exit(1)
	}
}

func baseDBConfig(config skyconfig.Configuration) skydb.DBConfig {
	passwordHistoryEnabled := config.UserAudit.PwHistorySize > 0 ||
		config.UserAudit.PwHistoryDays > 0

	return skydb.DBConfig{
		CanMigrate:             config.App.DevMode,
		PasswordHistoryEnabled: passwordHistoryEnabled,
	}
}

func ensureDB(config skyconfig.Configuration) func() (skydb.Conn, error) {
	logger := logging.LoggerEntryWithTag("main", "skydb")
	connOpener := func() (skydb.Conn, error) {
		return skydb.Open(
			context.Background(),
			config.DB.ImplName,
			config.App.Name,
			config.App.AccessControl,
			config.DB.Option,
			baseDBConfig(config),
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
		logger.Errorf("Failed to start skygear: %v", connError)
		if attempt >= 5 {
			logger.Fatalf("Failed to start skygear server because connection to database cannot be opened.")
		}

		logger.Info("Retrying in 1 second...")
		time.Sleep(time.Second * time.Duration(1))
	}
}

func initUserAuthRecordKeys(connOpener func() (skydb.Conn, error), authRecordKeys [][]string) {
	logger := logging.LoggerEntryWithTag("main", "auth")
	conn, err := connOpener()
	if err != nil {
		logger.Warnf("Failed to init user auth record keys: %v", err)
	}

	defer conn.Close()

	if err := conn.EnsureAuthRecordKeysExist(authRecordKeys); err != nil {
		panic(err)
	}

	if err := conn.EnsureAuthRecordKeysIndexesMatch(authRecordKeys); err != nil {
		panic(err)
	}
}

func initAssetStore(config skyconfig.Configuration) asset.Store {
	var store asset.Store
	switch config.AssetStore.ImplName {
	default:
		panic("unrecgonized asset store implementation: " + config.AssetStore.ImplName)
	case "fs":
		store = asset.NewFileStore(
			config.AssetStore.FileSystemStore.Path,
			config.AssetStore.FileSystemStore.URLPrefix,
			config.AssetStore.FileSystemStore.Secret,
			config.AssetStore.Public,
		)
	case "s3":
		s3Store, err := asset.NewS3Store(
			config.AssetStore.S3Store.AccessToken,
			config.AssetStore.S3Store.SecretToken,
			config.AssetStore.S3Store.Region,
			config.AssetStore.S3Store.Bucket,
			config.AssetStore.S3Store.URLPrefix,
			config.AssetStore.Public,
		)
		if err != nil {
			panic("failed to initialize asset.S3Store: " + err.Error())
		}
		store = s3Store
	case "cloud":
		cloudStore, err := asset.NewCloudStore(
			config.App.Name,
			config.AssetStore.CloudStore.Host,
			config.AssetStore.CloudStore.Token,
			config.AssetStore.CloudStore.PublicPrefix,
			config.AssetStore.CloudStore.PrivatePrefix,
			config.AssetStore.Public,
		)
		if err != nil {
			panic("Fail to initialize asset.CloudStore: " + err.Error())
		}
		store = cloudStore
	}
	return store
}

func initDevice(config skyconfig.Configuration, connOpener func() (skydb.Conn, error)) {
	logger := logging.LoggerEntryWithTag("main", "device")
	// TODO: Create a device service to check APNs to remove obsolete devices.
	// The current implementation deletes pubsub devices if the last registered
	// time is more than 1 day old.
	conn, err := connOpener()
	if err != nil {
		logger.Warnf("Failed to delete outdated devices: %v", err)
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
	if config.Baidu.Enable {
		baidu := initBaiduPusher(config)
		routeSender.Route("baidu-android", baidu)
	}
	return routeSender
}

func initAPNSPusher(config skyconfig.Configuration, connOpener func() (skydb.Conn, error)) push.APNSPusher {
	logger := logging.LoggerEntryWithTag("main", "push")
	var pushSender push.APNSPusher

	switch config.APNS.Type {
	case "cert":
		pushSender = initCertBasedAPNSPusher(config, connOpener)
	case "token":
		pushSender = initTokenBasedAPNSPusher(config, connOpener)
	default:
		logger.Fatalf("Unknown APNS Type: %s", config.APNS.Type)
	}

	go pushSender.Start()
	return pushSender
}

func initCertBasedAPNSPusher(
	config skyconfig.Configuration,
	connOpener func() (skydb.Conn, error),
) push.APNSPusher {
	logger := logging.LoggerEntryWithTag("main", "push")
	cert := config.APNS.CertConfig.Cert
	key := config.APNS.CertConfig.Key
	if config.APNS.CertConfig.Cert == "" && config.APNS.CertConfig.CertPath != "" {
		certPEMBlock, err := ioutil.ReadFile(config.APNS.CertConfig.CertPath)
		if err != nil {
			logger.Fatalf("Failed to load the APNS Cert: %v", err)
		}
		cert = string(certPEMBlock)
	}

	if config.APNS.CertConfig.Key == "" && config.APNS.CertConfig.KeyPath != "" {
		keyPEMBlock, err := ioutil.ReadFile(config.APNS.CertConfig.KeyPath)
		if err != nil {
			logger.Fatalf("Failed to load the APNS Key: %v", err)
		}
		key = string(keyPEMBlock)
	}

	pushSender, err := push.NewCertBasedAPNSPusher(
		connOpener,
		push.GatewayType(config.APNS.Env),
		cert,
		key,
	)
	if err != nil {
		logger.Fatalf("Failed to set up push sender: %v", err)
	}

	return pushSender
}

func initTokenBasedAPNSPusher(
	config skyconfig.Configuration,
	connOpener func() (skydb.Conn, error),
) push.APNSPusher {
	logger := logging.LoggerEntryWithTag("main", "push")
	key := config.APNS.TokenConfig.Key
	keyPath := config.APNS.TokenConfig.KeyPath
	if key == "" && keyPath != "" {
		keyBytes, err := ioutil.ReadFile(keyPath)
		if err != nil {
			logger.Fatalf("Failed to load APNS key: %v", err)
		}

		key = string(keyBytes)
	}

	pushSender, err := push.NewTokenBasedAPNSPusher(
		connOpener,
		push.GatewayType(config.APNS.Env),
		config.APNS.TokenConfig.TeamID,
		config.APNS.TokenConfig.KeyID,
		key,
	)
	if err != nil {
		logger.Fatalf("Failed to set up push sender: %v", err)
	}

	return pushSender
}

func initGCMPusher(config skyconfig.Configuration) *push.GCMPusher {
	return &push.GCMPusher{APIKey: config.GCM.APIKey}
}

func initBaiduPusher(config skyconfig.Configuration) *push.BaiduPusher {
	return push.NewBaiduPusher(config.Baidu.APIKey, config.Baidu.SecretKey)
}

func initSubscription(config skyconfig.Configuration, connOpener func() (skydb.Conn, error), hub *pubsub.Hub, pushSender push.Sender) {
	logger := logging.LoggerEntryWithTag("main", "subscription")
	notifiers := []subscription.Notifier{subscription.NewHubNotifier(hub)}
	if pushSender != nil {
		notifiers = append(notifiers, subscription.NewPushNotifier(pushSender))
	}

	subscriptionService := &subscription.Service{
		ConnOpener: connOpener,
		Notifier:   subscription.NewMultiNotifier(notifiers...),
	}
	logger.Infoln("Subscription Service listening...")
	go subscriptionService.Run()
}

func initPlugin(config skyconfig.Configuration, ctx *plugin.Context) {
	logger := logging.LoggerEntryWithTag("main", "logger")
	logger.Infof("Supported plugin transports: %s", strings.Join(plugin.SupportedTransports(), ", "))

	if ctx.Scheduler != nil {
		ctx.Scheduler.Start()
	}

	for _, pluginConfig := range config.Plugin {
		ctx.AddPluginConfiguration(pluginConfig.Transport, pluginConfig.Path, pluginConfig.Args)
	}

	ctx.InitPlugins()
}

func initLogger(config skyconfig.Configuration) {
	mainLogger := logging.LoggerEntryWithTag("main", "") // untagged logger

	// Create a sentry hook. This hook will be shared among all
	// loggers.
	var sentryHook logrus.Hook
	if config.LogHook.SentryDSN != "" {
		hook, err := newSentryHook(config)
		if err != nil {
			mainLogger.WithError(err).Panicf("Unable to configure sentry hook")
		}

		sentryHook = hook
	}

	// Create a function for configuring logger, this will be used
	// to configure the standard logger and configure all other loggers
	// created via the `logging` package.
	configure := func(name string, logger *logrus.Logger) {
		// Set stderr
		logger.Out = os.Stderr

		// Set global log level
		if level, err := logrus.ParseLevel(config.LOG.Level); err == nil {
			logger.Level = level
		} else {
			logger.Warnf("log: error parsing config: %v", err)
			logger.Warnln("log: fall back to `debug`")
			logger.Level = logrus.DebugLevel
		}

		// Set logger specific level
		sanitized := strings.Replace(strings.ToLower(name), ".", "_", -1)
		if loggerLevel, ok := config.LOG.LoggersLevel[sanitized]; ok {
			if level, err := logrus.ParseLevel(loggerLevel); err == nil {
				logger.Level = level
			}
		}

		// Set format
		var formatter logrus.Formatter
		if config.LOG.Formatter == "text" {
			formatter = &logging.TextFormatter{}
		} else if config.LOG.Formatter == "json" {
			formatter = &logrus.JSONFormatter{}
		} else {
			mainLogger.Warnf("log: Formatter '%s' is not defined, default to 'text'.", config.LOG.Formatter)
			formatter = &logging.TextFormatter{}
		}
		logger.Formatter = formatter

		// Set sentry hook
		if sentryHook != nil {
			logger.Hooks.Add(sentryHook)
		}
	}

	// Register the function, and the update the standard logger and all
	// loggers configured before.
	logging.SetConfigureLoggerHandler(configure)
	configure("", logrus.StandardLogger())
	for loggerName, logger := range logging.Loggers() {
		configure(loggerName, logger)
	}

	// Configure audit handlers.
	err := audit.InitTrailHandler(config.UserAudit.Enabled, config.UserAudit.TrailHandlerURL)
	if err != nil {
		mainLogger.Fatalf("user-audit: error when initializing trail handler %v", err)
		return
	}
}

func higherLogLevels(minLevel logrus.Level) []logrus.Level {
	levels := []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}

	output := make([]logrus.Level, 0, len(levels))
	for _, level := range levels {
		if level <= minLevel {
			output = append(output, level)
		}
	}
	return output
}

func newSentryHook(config skyconfig.Configuration) (logrus.Hook, error) {
	level, err := logrus.ParseLevel(config.LogHook.SentryLevel)
	if err != nil {
		return nil, fmt.Errorf("log-hook: error parsing sentry-level: %v", err)
	}

	levels := higherLogLevels(level)
	tags := map[string]string{
		"version": skyversion.Version(),
	}

	hook, err := logrus_sentry.NewWithTagsSentryHook(
		config.LogHook.SentryDSN,
		tags,
		levels)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize Sentry: %v", err)
	}
	hook.Timeout = 1 * time.Second
	return hook, nil
}
