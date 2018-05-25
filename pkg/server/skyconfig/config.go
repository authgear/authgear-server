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

package skyconfig

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/skygeario/skygear-server/pkg/server/uuid"
)

// parseBool parses a string representation of boolean value into a boolean
// type. In addition to what strconv.ParseBool recognize, it also recognize
// Yes, yes, YES, y, No, no, NO, n. Any other value returns an error.
func parseBool(str string) (bool, error) {
	switch str {
	case "Yes", "yes", "YES", "y":
		return true, nil
	case "No", "no", "NO", "n":
		return false, nil
	default:
		return strconv.ParseBool(str)
	}
}

// parseCommaSeparatedString parses a string representation of a comma separated list
// into a slice of string, omitting empty strings.
func parseCommaSeparatedString(str string) []string {
	splits := strings.Split(str, ",")
	results := make([]string, 0, len(splits))
	for _, split := range splits {
		split = strings.TrimSpace(split)
		if split != "" {
			results = append(results, split)
		}
	}
	return results
}

// parseAuthRecordKeys parses a string representation of a comma separated list
// to keys and key tuples
//
// example:
// a,b,c => [[a], [b], [c]]
// (a),(b,c),(d,e,f) => [[a], [b,c], [d,e,f]]
//
// error example:
// a,(b,(d),c)
// a,a(b)c
func parseAuthRecordKeys(str string) ([][]string, error) {
	if str == "" {
		return [][]string{}, fmt.Errorf("Empty string")
	}

	splits := strings.Split(str, ",")
	results := [][]string{}
	container := []string{}
	level := 0
	for _, split := range splits {
		split = strings.TrimSpace(split)
		content := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(split, "("), ")"))

		isGroupOpening := strings.HasPrefix(split, "(")
		isGroupClosing := strings.HasSuffix(split, ")")

		// validation
		if strings.Contains(content, "(") || strings.Contains(content, ")") || (level > 0 && isGroupOpening) {
			return [][]string{}, fmt.Errorf("Unexpected char in " + content)
		}

		if isGroupOpening {
			container = []string{}
			level++
		}

		container = append(container, content)

		if isGroupClosing {
			level--
			sort.Strings(container)
			results = append(results, container)
		}

		if !isGroupOpening && !isGroupClosing && level == 0 {
			results = append(results, container)
			container = []string{}
		}
	}

	return results, nil
}

type PluginConfig struct {
	Transport string
	Path      string
	Args      []string
}

// Configuration is Skygear's configuration
// The configuration will load in following order:
// 1. The ENV
// 2. The key/value in .env file
// 3. The config in *.ini (To-be depreacted)
type Configuration struct {
	HTTP struct {
		Host string `json:"host"`
	} `json:"http"`
	App struct {
		Name            string     `json:"name"`
		APIKey          string     `json:"api_key"`
		MasterKey       string     `json:"master_key"`
		AccessControl   string     `json:"access_control"`
		AuthRecordKeys  [][]string `json:"auth_record_keys"`
		DevMode         bool       `json:"dev_mode"`
		CORSHost        string     `json:"cors_host"`
		Slave           bool       `json:"slave"`
		ResponseTimeout int64      `json:"response_timeout"`
	} `json:"app"`
	DB struct {
		ImplName string `json:"implementation"`
		Option   string `json:"option"`
	} `json:"database"`
	TokenStore struct {
		ImplName string `json:"implementation"`
		Path     string `json:"path"`
		Prefix   string `json:"prefix"`
		Expiry   int64  `json:"expiry"`
		Secret   string `json:"secret"`
	} `json:"-"`
	Auth struct {
		CustomTokenSecret string `json:"custom_token_secret"`
	} `json:"auth"`
	AssetStore struct {
		ImplName string `json:"implementation"`
		Public   bool   `json:"public"`

		FileSystemStore struct {
			Path      string `json:"-"`
			URLPrefix string `json:"url_prefix"`
			Secret    string `json:"secret"`
		} `json:"fs"`

		S3Store struct {
			AccessToken string `json:"access_key"`
			SecretToken string `json:"secret_key"`
			Region      string `json:"region"`
			Bucket      string `json:"bucket"`
			URLPrefix   string `json:"url_prefix"`
		} `json:"s3"`

		CloudStore struct {
			Host          string `json:"host"`
			Token         string `json:"token"`
			PublicPrefix  string `json:"public_prefix"`
			PrivatePrefix string `json:"private_prefix"`
		} `json:"cloud"`
	} `json:"asset_store"`
	APNS struct {
		Enable bool   `json:"enable"`
		Type   string `json:"type"`
		Env    string `json:"env"`

		CertConfig struct {
			Cert     string `json:"cert"`
			Key      string `json:"key"`
			CertPath string `json:"-"`
			KeyPath  string `json:"-"`
		} `json:"cert_config"`

		TokenConfig struct {
			TeamID  string `json:"team_id"`
			KeyID   string `json:"key_id"`
			Key     string `json:"key"`
			KeyPath string `json:"-"`
		} `json:"token_config"`
	} `json:"apns"`
	GCM struct {
		Enable bool   `json:"enable"`
		APIKey string `json:"api_key"`
	} `json:"gcm"`
	Baidu struct {
		Enable    bool   `json:"enable"`
		APIKey    string `json:"api_key"`
		SecretKey string `json:"secret_key"`
	} `json:"baidu"`
	LOG struct {
		Level           string            `json:"-"`
		LoggersLevel    map[string]string `json:"-"`
		RouterByteLimit int64             `json:"-"`
		Formatter       string            `json:"-"`
	} `json:"log"`
	LogHook struct {
		SentryDSN   string
		SentryLevel string
	} `json:"-"`
	Zmq struct {
		Timeout   int `json:"timeout"`
		MaxBounce int `json:"max_bounce"`
	} `json:"zmq"`
	Plugin    map[string]*PluginConfig `json:"-"`
	UserAudit struct {
		Enabled             bool     `json:"enabled"`
		TrailHandlerURL     string   `json:"trail_handler_url"`
		PwMinLength         int      `json:"pw_min_length"`
		PwUppercaseRequired bool     `json:"pw_uppercase_required"`
		PwLowercaseRequired bool     `json:"pw_lowercase_required"`
		PwDigitRequired     bool     `json:"pw_digit_required"`
		PwSymbolRequired    bool     `json:"pw_symbol_required"`
		PwMinGuessableLevel int      `json:"pw_min_guessable_level"`
		PwExcludedKeywords  []string `json:"pw_excluded_keywords"`
		PwExcludedFields    []string `json:"pw_excluded_fields"`
		PwHistorySize       int      `json:"pw_history_size"`
		PwHistoryDays       int      `json:"pw_history_days"`
		PwExpiryDays        int      `json:"pw_expiry_days"`
	} `json:"user_audit"`
	Verification struct {
		Required bool `json:"required"`
	} `json:"verification"`
}

func NewConfiguration() Configuration {
	config := Configuration{}
	config.HTTP.Host = ":3000"
	config.App.Name = "myapp"
	config.App.AccessControl = "role"
	config.App.AuthRecordKeys = [][]string{[]string{"username"}, []string{"email"}}
	config.App.DevMode = true
	config.App.CORSHost = "*"
	config.App.Slave = false
	config.App.ResponseTimeout = 60
	config.DB.ImplName = "pq"
	config.DB.Option = "postgres://postgres:@localhost/postgres?sslmode=disable"
	config.TokenStore.ImplName = "fs"
	config.TokenStore.Path = "data/token"
	config.TokenStore.Expiry = 0
	config.AssetStore.ImplName = "fs"
	config.AssetStore.FileSystemStore.Path = "data/asset"
	config.AssetStore.FileSystemStore.URLPrefix = "http://localhost:3000/files"
	config.APNS.Enable = false
	config.APNS.Type = "cert"
	config.APNS.Env = "sandbox"
	config.GCM.Enable = false
	config.Baidu.Enable = false
	config.LOG.Level = "debug"
	config.LOG.LoggersLevel = map[string]string{
		"plugin": "info",
	}
	config.LOG.RouterByteLimit = 100000
	config.LOG.Formatter = "text"
	config.LogHook.SentryLevel = "error"
	config.Zmq.Timeout = 30
	config.Zmq.MaxBounce = 10
	config.Plugin = map[string]*PluginConfig{}
	return config
}

func NewConfigurationWithKeys() Configuration {
	config := NewConfiguration()
	config.App.APIKey = uuid.New()
	config.App.MasterKey = uuid.New()
	return config
}

func (config *Configuration) Validate() error {
	if config.App.Name == "" {
		return errors.New("APP_NAME is not set")
	}
	if config.App.APIKey == "" {
		return errors.New("API_KEY is not set")
	}
	if config.App.MasterKey == "" {
		return errors.New("MASTER_KEY is not set")
	}
	if config.App.APIKey == config.App.MasterKey {
		return errors.New("MASTER_KEY cannot be the same as API_KEY")
	}
	if !regexp.MustCompile("^[A-Za-z0-9_]+$").MatchString(config.App.Name) {
		return fmt.Errorf("APP_NAME '%s' contains invalid characters other than alphanumerics or underscores", config.App.Name)
	}
	if config.APNS.Enable && !regexp.MustCompile("^(sandbox|production)$").MatchString(config.APNS.Env) {
		return fmt.Errorf("APNS_ENV must be sandbox or production")
	}
	if config.APNS.Enable && !regexp.MustCompile("^(cert|token)$").MatchString(config.APNS.Type) {
		return fmt.Errorf("APNS_TYPE must be cert or token")
	}
	return config.checkAuthRecordKeysDuplication()
}

func (config *Configuration) checkAuthRecordKeysDuplication() error {
	check := map[string]interface{}{}
	for _, result := range config.App.AuthRecordKeys {
		c := strings.Join(result, ",")
		if _, found := check[c]; found {
			return fmt.Errorf("AUTH_RECORD_KEYS cannot have duplicated keys '%s'", c)
		}
		check[c] = result
	}

	return nil
}

// ReadFromEnv reads from environment variable and update the configuration.
// nolint: gocyclo
func (config *Configuration) ReadFromEnv() {
	envErr := godotenv.Load()
	if envErr != nil {
		log.Print("Error in loading .env file")
	}

	config.readHost()

	appAPIKey := os.Getenv("API_KEY")
	if appAPIKey != "" {
		config.App.APIKey = appAPIKey
	}

	appMasterKey := os.Getenv("MASTER_KEY")
	if appMasterKey != "" {
		config.App.MasterKey = appMasterKey
	}

	appName := os.Getenv("APP_NAME")
	if appName != "" {
		config.App.Name = appName
	}

	corsHost := os.Getenv("CORS_HOST")
	if corsHost != "" {
		config.App.CORSHost = corsHost
	}

	accessControl := os.Getenv("ACCESS_CONRTOL")
	if accessControl != "" {
		config.App.AccessControl = accessControl
	}

	if authRecordKeys, err := parseAuthRecordKeys(os.Getenv("AUTH_RECORD_KEYS")); err == nil {
		config.App.AuthRecordKeys = authRecordKeys
	}

	if devMode, err := parseBool(os.Getenv("DEV_MODE")); err == nil {
		config.App.DevMode = devMode
	}

	dbImplName := os.Getenv("DB_IMPL_NAME")
	if dbImplName != "" {
		config.DB.ImplName = dbImplName
	}

	if config.DB.ImplName == "pq" && os.Getenv("DATABASE_URL") != "" {
		config.DB.Option = os.Getenv("DATABASE_URL")
	}

	if slave, err := parseBool(os.Getenv("SLAVE")); err == nil {
		config.App.Slave = slave
	}

	if timeout, err := strconv.ParseInt(os.Getenv("RESPONSE_TIMEOUT"), 10, 64); err == nil {
		config.App.ResponseTimeout = timeout
	}

	if bounceCount, err := strconv.ParseInt(os.Getenv("ZMQ_MAX_BOUNCE"), 10, 0); err == nil {
		config.Zmq.MaxBounce = int(bounceCount)
	}

	if secret := os.Getenv("CUSTOM_TOKEN_SECRET"); secret != "" {
		config.Auth.CustomTokenSecret = secret
	}

	config.readTokenStore()
	config.readAssetStore()
	config.readAPNS()
	config.readGCM()
	config.readBaidu()
	config.readLog()
	config.readPlugins()
	config.readUserAudit()
	config.readUserVerification()
}

func (config *Configuration) readHost() {
	// Default to :3000 if both HOST and PORT is missing
	host := os.Getenv("HOST")
	if host != "" {
		config.HTTP.Host = host
	}
	if config.HTTP.Host == "" {
		port := os.Getenv("PORT")
		if port != "" {
			config.HTTP.Host = ":" + port
		}
	}
}

func (config *Configuration) readTokenStore() {
	tokenStore := os.Getenv("TOKEN_STORE")
	if tokenStore != "" {
		config.TokenStore.ImplName = tokenStore
	}
	tokenStorePath := os.Getenv("TOKEN_STORE_PATH")
	if tokenStorePath != "" {
		config.TokenStore.Path = tokenStorePath
	}

	tokenStorePrefix := os.Getenv("TOKEN_STORE_PREFIX")
	if tokenStorePrefix != "" {
		config.TokenStore.Prefix = tokenStorePrefix
	}

	if expiry, err := strconv.ParseInt(os.Getenv("TOKEN_STORE_EXPIRY"), 10, 64); err == nil {
		config.TokenStore.Expiry = expiry
	}

	tokenStoreSecret := os.Getenv("TOKEN_STORE_SECRET")
	if tokenStoreSecret != "" {
		config.TokenStore.Secret = tokenStoreSecret
	} else {
		config.TokenStore.Secret = config.App.MasterKey
	}
}

func (config *Configuration) readAssetStore() {
	assetStore := os.Getenv("ASSET_STORE")
	if assetStore != "" {
		config.AssetStore.ImplName = assetStore
	}

	if assetStorePublic, err := parseBool(os.Getenv("ASSET_STORE_PUBLIC")); err == nil {
		config.AssetStore.Public = assetStorePublic
	}

	// Local Storage related
	assetStorePath := os.Getenv("ASSET_STORE_PATH")
	if assetStorePath != "" {
		config.AssetStore.FileSystemStore.Path = assetStorePath
	}
	assetStorePrefix := os.Getenv("ASSET_STORE_URL_PREFIX")
	if assetStorePrefix != "" {
		config.AssetStore.FileSystemStore.URLPrefix = assetStorePrefix
	}
	assetStoreSecret := os.Getenv("ASSET_STORE_SECRET")
	if assetStoreSecret != "" {
		config.AssetStore.FileSystemStore.Secret = assetStoreSecret
	}

	// S3 related
	assetStoreAccessKey := os.Getenv("ASSET_STORE_ACCESS_KEY")
	if assetStoreAccessKey != "" {
		config.AssetStore.S3Store.AccessToken = assetStoreAccessKey
	}
	assetStoreSecretKey := os.Getenv("ASSET_STORE_SECRET_KEY")
	if assetStoreSecretKey != "" {
		config.AssetStore.S3Store.SecretToken = assetStoreSecretKey
	}
	assetStoreRegion := os.Getenv("ASSET_STORE_REGION")
	if assetStoreRegion != "" {
		config.AssetStore.S3Store.Region = assetStoreRegion
	}
	assetStoreBucket := os.Getenv("ASSET_STORE_BUCKET")
	if assetStoreBucket != "" {
		config.AssetStore.S3Store.Bucket = assetStoreBucket
	}
	assetStoreS3URLPrefix := os.Getenv("ASSET_STORE_S3_URL_PREFIX")
	if assetStoreS3URLPrefix != "" {
		config.AssetStore.S3Store.URLPrefix = assetStoreS3URLPrefix
	}

	// Cloud Asset related
	cloudAssetHost := os.Getenv("CLOUD_ASSET_HOST")
	if cloudAssetHost != "" {
		config.AssetStore.CloudStore.Host = cloudAssetHost
	}
	cloudAssetToken := os.Getenv("CLOUD_ASSET_TOKEN")
	if cloudAssetToken != "" {
		config.AssetStore.CloudStore.Token = cloudAssetToken
	}
	cloudAssetPublicPrefix := os.Getenv("CLOUD_ASSET_PUBLIC_PREFIX")
	if cloudAssetPublicPrefix != "" {
		config.AssetStore.CloudStore.PublicPrefix = cloudAssetPublicPrefix
	}
	cloudAssetPrivatePrefix := os.Getenv("CLOUD_ASSET_PRIVATE_PREFIX")
	if cloudAssetPrivatePrefix != "" {
		config.AssetStore.CloudStore.PrivatePrefix = cloudAssetPrivatePrefix
	}
}

func (config *Configuration) readAPNS() {
	if shouldEnableAPNS, err := parseBool(os.Getenv("APNS_ENABLE")); err == nil {
		config.APNS.Enable = shouldEnableAPNS
	}

	if !config.APNS.Enable {
		return
	}

	env := os.Getenv("APNS_ENV")
	if env != "" {
		config.APNS.Env = env
	}

	apnsType := os.Getenv("APNS_TYPE")
	if apnsType != "" {
		config.APNS.Type = apnsType
	}

	switch strings.ToLower(config.APNS.Type) {
	case "cert":
		config.readAPNSCert()
	case "token":
		config.readAPNSToken()
	}
}

func (config *Configuration) readAPNSCert() {
	cert, key := os.Getenv("APNS_CERTIFICATE"), os.Getenv("APNS_PRIVATE_KEY")
	if cert != "" {
		config.APNS.CertConfig.Cert = cert
	}
	if key != "" {
		config.APNS.CertConfig.Key = key
	}

	certPath, keyPath := os.Getenv("APNS_CERTIFICATE_PATH"), os.Getenv("APNS_PRIVATE_KEY_PATH")
	if certPath != "" {
		config.APNS.CertConfig.CertPath = certPath
	}
	if keyPath != "" {
		config.APNS.CertConfig.KeyPath = keyPath
	}
}

func (config *Configuration) readAPNSToken() {
	teamID := os.Getenv("APNS_TEAM_ID")
	if teamID != "" {
		config.APNS.TokenConfig.TeamID = teamID
	}

	keyID := os.Getenv("APNS_KEY_ID")
	if keyID != "" {
		config.APNS.TokenConfig.KeyID = keyID
	}

	key := os.Getenv("APNS_TOKEN_KEY")
	if key != "" {
		config.APNS.TokenConfig.Key = key
	}

	keyPath := os.Getenv("APNS_TOKEN_KEY_PATH")
	if keyPath != "" {
		config.APNS.TokenConfig.KeyPath = keyPath
	}
}

func (config *Configuration) readGCM() {
	if shouldEnableGCM, err := parseBool(os.Getenv("GCM_ENABLE")); err == nil {
		config.GCM.Enable = shouldEnableGCM
	}

	gcmAPIKey := os.Getenv("GCM_APIKEY")
	if gcmAPIKey != "" {
		config.GCM.APIKey = gcmAPIKey
	}
}

func (config *Configuration) readBaidu() {
	if shouldEnableBaidu, err := parseBool(os.Getenv("BAIDU_ENABLE")); err == nil {
		config.Baidu.Enable = shouldEnableBaidu
	}

	baiduAPIKey := os.Getenv("BAIDU_API_KEY")
	if baiduAPIKey != "" {
		config.Baidu.APIKey = baiduAPIKey
	}

	baiduSecretKey := os.Getenv("BAIDU_SECRET_KEY")
	if baiduSecretKey != "" {
		config.Baidu.SecretKey = baiduSecretKey
	}
}

func (config *Configuration) readLog() {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel != "" {
		config.LOG.Level = logLevel
	}

	for _, environ := range os.Environ() {
		if !strings.HasPrefix(environ, "LOG_LEVEL_") {
			continue
		}

		components := strings.SplitN(environ, "=", 2)
		loggerName := strings.ToLower(strings.TrimPrefix(components[0], "LOG_LEVEL_"))
		loggerLevel := components[1]
		config.LOG.LoggersLevel[loggerName] = loggerLevel
	}

	if byteLimit, err := strconv.ParseInt(os.Getenv("LOG_ROUTER_BYTE_LIMIT"), 10, 64); err == nil {
		config.LOG.RouterByteLimit = byteLimit
	}

	sentry := os.Getenv("SENTRY_DSN")
	if sentry != "" {
		config.LogHook.SentryDSN = sentry
	}

	sentryLevel := os.Getenv("SENTRY_LEVEL")
	if sentryLevel != "" {
		config.LogHook.SentryLevel = sentryLevel
	}

	if formatter := os.Getenv("LOG_FORMATTER"); formatter != "" {
		config.LOG.Formatter = formatter
	}
}

func (config *Configuration) readPlugins() {
	timeoutStr := os.Getenv("ZMQ_TIMEOUT")
	timeout, err := strconv.Atoi(timeoutStr)
	if err == nil {
		config.Zmq.Timeout = timeout
	}

	plugin := os.Getenv("PLUGINS")
	if plugin == "" {
		return
	}

	plugins := strings.Split(plugin, ",")
	for _, p := range plugins {
		pluginConfig := &PluginConfig{}
		pluginConfig.Transport = os.Getenv(p + "_TRANSPORT")
		pluginConfig.Path = os.Getenv(p + "_PATH")
		args := os.Getenv(p + "_ARGS")
		if args != "" {
			pluginConfig.Args = strings.Split(args, ",")
		}
		config.Plugin[p] = pluginConfig
	}
}

// nolint: gocyclo
func (config *Configuration) readUserAudit() {
	if v, err := parseBool(os.Getenv("USER_AUDIT_ENABLED")); err == nil {
		config.UserAudit.Enabled = v
	}
	config.UserAudit.TrailHandlerURL = os.Getenv("USER_AUDIT_TRAIL_HANDLER_URL")
	if v, err := strconv.ParseInt(os.Getenv("USER_AUDIT_PW_MIN_LENGTH"), 10, 0); err == nil && v > 0 {
		config.UserAudit.PwMinLength = int(v)
	}
	if v, err := parseBool(os.Getenv("USER_AUDIT_PW_UPPERCASE_REQUIRED")); err == nil {
		config.UserAudit.PwUppercaseRequired = v
	}
	if v, err := parseBool(os.Getenv("USER_AUDIT_PW_LOWERCASE_REQUIRED")); err == nil {
		config.UserAudit.PwLowercaseRequired = v
	}
	if v, err := parseBool(os.Getenv("USER_AUDIT_PW_DIGIT_REQUIRED")); err == nil {
		config.UserAudit.PwDigitRequired = v
	}
	if v, err := parseBool(os.Getenv("USER_AUDIT_PW_SYMBOL_REQUIRED")); err == nil {
		config.UserAudit.PwSymbolRequired = v
	}
	if v, err := strconv.ParseInt(os.Getenv("USER_AUDIT_PW_MIN_GUESSABLE_LEVEL"), 10, 0); err == nil && v > 0 && v <= 5 {
		config.UserAudit.PwMinGuessableLevel = int(v)
	}
	if v := parseCommaSeparatedString(os.Getenv("USER_AUDIT_PW_EXCLUDED_KEYWORDS")); len(v) > 0 {
		config.UserAudit.PwExcludedKeywords = v
	}
	if v := parseCommaSeparatedString(os.Getenv("USER_AUDIT_PW_EXCLUDED_FIELDS")); len(v) > 0 {
		config.UserAudit.PwExcludedFields = v
	}
	if v, err := strconv.ParseInt(os.Getenv("USER_AUDIT_PW_HISTORY_SIZE"), 10, 0); err == nil && v > 0 {
		config.UserAudit.PwHistorySize = int(v)
	}
	if v, err := strconv.ParseInt(os.Getenv("USER_AUDIT_PW_HISTORY_DAYS"), 10, 0); err == nil && v > 0 {
		config.UserAudit.PwHistoryDays = int(v)
	}
	if v, err := strconv.ParseInt(os.Getenv("USER_AUDIT_PW_EXPIRY_DAYS"), 10, 0); err == nil && v > 0 {
		config.UserAudit.PwExpiryDays = int(v)
	}
}

func (config *Configuration) readUserVerification() {
	if v, err := parseBool(os.Getenv("VERIFY_REQUIRED")); err == nil {
		config.Verification.Required = v
	}
}
