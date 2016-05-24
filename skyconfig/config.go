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
	"io/ioutil"
	"os"
	"regexp"

	"github.com/joho/godotenv"
	"github.com/oursky/gcfg"
	"github.com/skygeario/skygear-server/uuid"
)

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
		Name          string `json:"name"`
		APIKey        string `gcfg:"api-key" json:"api_key"`
		MasterKey     string `gcfg:"master-key" json:"master_key"`
		AccessControl string `gcfg:"access-control" json:"access_control"`
		DevMode       bool   `gcfg:"dev-mode" json:"dev_mode"`
		CORSHost      string `gcfg:"cors-host" json:"cors_host"`
	} `json:"app"`
	DB struct {
		ImplName string `gcfg:"implementation" json:"implementation"`
		Option   string `json:"option"`
	} `json:"database"`
	TokenStore struct {
		ImplName string `gcfg:"implementation" json:"implementation"`
		Path     string `gcfg:"path" json:"path"`
		Prefix   string `gcfg:"prefix" json:"prefix"`
	} `gcfg:"token-store" json:"-"`
	AssetStore struct {
		ImplName string `gcfg:"implementation" json:"implementation"`
		Public   bool   `json:"public"`

		// followings only used when ImplName = fs
		Path string `json:"-"`

		// followings only used when ImplName = s3
		AccessToken string `gcfg:"access-key" json:"access_key"`
		SecretToken string `gcfg:"secret-key" json:"secret_key"`
		Region      string `json:"region"`
		Bucket      string `json:"bucket"`
	} `gcfg:"asset-store" json:"asset_store"`
	AssetURLSigner struct {
		URLPrefix string `gcfg:"url-prefix" json:"url_prefix"`
		Secret    string `json:"secret"`
	} `gcfg:"asset-url-signer" json:"asset_signer"`
	APNS struct {
		Enable   bool   `json:"enable"`
		Env      string `json:"env"`
		Cert     string `json:"cert"`
		Key      string `json:"key"`
		CertPath string `gcfg:"cert-path" json:"-"`
		KeyPath  string `gcfg:"key-path" json:"-"`
	} `json:"apns"`
	GCM struct {
		Enable bool   `json:"enable"`
		APIKey string `gcfg:"api-key" json:"api_key"`
	} `json:"gcm"`
	LOG struct {
		Level string `json:"-"`
	} `json:"log"`
	LogHook struct {
		SentryDSN   string `gcfg:"sentry-dsn"`
		SentryLevel string `gcfg:"sentry-level"`
	} `gcfg:"log-hook" json:"-"`
	Plugin map[string]*struct {
		Transport string
		Path      string
		Args      []string
	} `json:"-"`
	// the alembic section here is to make the config be parsed correctly
	// the values should not be used
	UselessAlembic struct {
		ScriptLocation string `gcfg:"script_location"`
	} `gcfg:"alembic" json:"-"`
}

func NewConfiguration() Configuration {
	config := Configuration{}
	config.HTTP.Host = ":3000"
	config.App.Name = "myapp"
	config.App.AccessControl = "role"
	config.App.DevMode = true
	config.DB.ImplName = "pq"
	config.DB.Option = "postgres://postgres:@localhost/postgres?sslmode=disable"
	config.TokenStore.ImplName = "fs"
	config.TokenStore.Path = "data/token"
	config.AssetStore.ImplName = "fs"
	config.AssetStore.Path = "data/asset"
	config.AssetURLSigner.URLPrefix = "http://localhost:3000/files"
	config.APNS.Enable = false
	config.APNS.Env = "sandbox"
	config.GCM.Enable = false
	config.LOG.Level = "debug"
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
	if !regexp.MustCompile("^[A-Za-z0-9_]+$").MatchString(config.App.Name) {
		return fmt.Errorf("APP_NAME '%s' contains invalid characters other than alphanumberics or underscores", config.App.Name)
	}
	if config.APNS.Enable && !regexp.MustCompile("^(sandbox|production)$").MatchString(config.APNS.Env) {
		return fmt.Errorf("APNS_ENV must be sandbox or production")
	}
	return nil
}

// ReadFromIni reads a configuration from file specified by path
func (config *Configuration) ReadFromINI(path string) error {
	if err := gcfg.ReadFileInto(config, path); err != nil {
		return err
	}
	return config.Validate()
}

func (config *Configuration) ReadFromEnv() error {
	envErr := godotenv.Load()
	if envErr != nil {
		fmt.Errorf("Error loading .env file")
	}

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
	if accessControl == "" {
		config.App.AccessControl = accessControl
	}

	DevMode := os.Getenv("DEV_MODE")
	if DevMode == "YES" {
		config.App.DevMode = true
	}
	if DevMode == "NO" {
		config.App.DevMode = false
	}

	dbImplName := os.Getenv("DB_IMPL_NAME")
	if dbImplName != "" {
		config.DB.ImplName = dbImplName
	}

	if config.DB.ImplName == "pq" && os.Getenv("DATABASE_URL") != "" {
		config.DB.Option = os.Getenv("DATABASE_URL")
	}

	tokenStorePrefix := os.Getenv("TOKEN_STORE_PREFIX")
	if tokenStorePrefix != "" {
		config.TokenStore.Prefix = tokenStorePrefix
	}

	shouldEnableAPNS := os.Getenv("APNS_ENABLE")
	if shouldEnableAPNS != "" {
		config.APNS.Enable = shouldEnableAPNS == "1" || shouldEnableAPNS == "YES"
	}

	env := os.Getenv("APNS_ENV")
	if env != "" {
		config.APNS.Env = env
	}

	err := readAPNS(config)
	if err != nil {
		return err
	}

	shouldEnableGCM := os.Getenv("GCM_ENABLE")
	if shouldEnableGCM != "" {
		config.GCM.Enable = shouldEnableGCM == "1" || shouldEnableGCM == "YES"
	}

	gcmAPIKey := os.Getenv("GCM_APIKEY")
	if gcmAPIKey != "" {
		config.GCM.APIKey = gcmAPIKey
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel != "" {
		config.LOG.Level = logLevel
	}

	sentry := os.Getenv("SENTRY_DSN")
	if sentry != "" {
		config.LogHook.SentryDSN = sentry
	}

	sentryLevel := os.Getenv("SENTRY_LEVEL")
	if sentryLevel != "" {
		config.LogHook.SentryLevel = logLevel
	}

	return nil
}

func readAPNS(config *Configuration) error {
	if !config.APNS.Enable {
		return nil
	}

	cert, key := os.Getenv("APNS_CERTIFICATE"), os.Getenv("APNS_PRIVATE_KEY")
	if cert != "" {
		config.APNS.Cert = cert
	}
	if key != "" {
		config.APNS.Key = key
	}

	certPath, keyPath := os.Getenv("APNS_CERTIFICATE_PATH"), os.Getenv("APNS_PRIVATE_KEY_PATH")
	if certPath != "" {
		config.APNS.CertPath = certPath
	}
	if keyPath != "" {
		config.APNS.KeyPath = keyPath
	}

	if config.APNS.Cert == "" && config.APNS.CertPath != "" {
		certPEMBlock, err := ioutil.ReadFile(config.APNS.CertPath)
		if err != nil {
			return err
		}
		config.APNS.Cert = string(certPEMBlock)
	}

	if config.APNS.Key == "" && config.APNS.KeyPath != "" {
		keyPEMBlock, err := ioutil.ReadFile(config.APNS.KeyPath)
		if err != nil {
			return err
		}
		config.APNS.Key = string(keyPEMBlock)
	}

	return nil
}
