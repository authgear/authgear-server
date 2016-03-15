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

	"github.com/oursky/gcfg"
)

// Configuration is Skygear's configuration
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

// ReadFileInto reads a configuration from file specified by path
func ReadFileInto(config *Configuration, path string) error {
	if err := gcfg.ReadFileInto(config, path); err != nil {
		return err
	}
	if config.HTTP.Host == "" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "3000"
		}
		config.HTTP.Host = ":" + port
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
	if config.App.Name == "" {
		return errors.New("app name is not set")
	}
	if !regexp.MustCompile("^[A-Za-z0-9_]+$").MatchString(config.App.Name) {
		return fmt.Errorf("app name '%s' contains invalid characters other than alphanumberics or underscores", config.App.Name)
	}

	corsHost := os.Getenv("CORS_HOST")
	if corsHost != "" {
		config.App.CORSHost = corsHost
	}

	if config.App.AccessControl == "" {
		config.App.AccessControl = "role"
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

	err := readAPNS(config)
	if err != nil {
		return err
	}
	return nil
}

func readAPNS(config *Configuration) error {
	shouldEnableAPNS := os.Getenv("APNS_ENABLE")
	if shouldEnableAPNS != "" {
		config.APNS.Enable = shouldEnableAPNS == "1"
	}
	if !config.APNS.Enable {
		return nil
	}

	env := os.Getenv("APNS_ENV")
	if env != "" {
		config.APNS.Env = env
	}

	cert, key := os.Getenv("APNS_CERTIFICATE"), os.Getenv("APNS_PRIVATE_KEY")
	if cert != "" {
		config.APNS.Cert = cert
	}
	if key != "" {
		config.APNS.Key = key
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
