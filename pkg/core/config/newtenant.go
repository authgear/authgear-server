package config

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/name"
)

type TenantConfiguration struct {
	Version    string            `json:"version" yaml:"version" msg:"version"`
	AppName    string            `json:"name" yaml:"name" msg:"name"`
	AppConfig  AppConfiguration  `json:"app" yaml:"app" msg:"app"`
	UserConfig UserConfiguration `json:"user" yaml:"user" msg:"user"`
	Hooks      []Hook            `json:"hooks" yaml:"hooks" msg:"hooks"`
}

type Hook struct {
	Async   bool   `json:"async" yaml:"async" msg:"async"`
	Event   string `json:"event" yaml:"event" msg:"event"`
	URL     string `json:"url" yaml:"url" msg:"url"`
	Timeout int    `json:"timeout" yaml:"timeout" msg:"timeout"`
}

func NewTenantConfiguration() TenantConfiguration {
	return TenantConfiguration{
		Version: "1",
		AppConfig: AppConfiguration{
			Version:     "1",
			DatabaseURL: "postgres://postgres:@localhost/postgres?sslmode=disable",
			SMTP: NewSMTPConfiguration{
				Port: 25,
				Mode: "normal",
			},
		},
		UserConfig: UserConfiguration{
			Version: "1",
			CORS: CORSConfiguration{
				Origin: "*",
			},
			Auth: NewAuthConfiguration{
				LoginIDKeys: []string{},
			},
			ForgotPassword: NewForgotPasswordConfiguration{
				SecureMatch:      false,
				Sender:           "no-reply@skygeario.com",
				Subject:          "Reset password instruction",
				ResetURLLifetime: 43200,
			},
			WelcomeEmail: NewWelcomeEmailConfiguration{
				Enabled: false,
				Sender:  "no-reply@skygeario.com",
				Subject: "Welcome!",
			},
			SSO: NewSSOConfiguration{
				JSSDKCDNURL: "https://code.skygear.io/js/skygear/latest/skygear.min.js",
			},
		},
	}
}

func NewTenantConfigurationFromEnv(_ *http.Request) (TenantConfiguration, error) {
	// TODO: Remove this function
	// Instead, we should call NewTenantConfigurationFromYAML once
	// and consistently return the same configuration for every incoming request
	return TenantConfiguration{}, errors.New("NewTenantConfigurationFromEnv")
}

func NewTenantConfigurationFromYAML(r io.Reader) (TenantConfiguration, error) {
	// TODO: Load tenant config at filepath
	return TenantConfiguration{}, errors.New("NewTenantConfigurationFromYAML")
}

func (c *TenantConfiguration) GetSSOProviderByName(name string) (SSOProviderConfiguration, bool) {
	for _, provider := range c.UserConfig.SSO.Providers {
		if provider.Name == name {
			return provider, true
		}
	}
	return SSOProviderConfiguration{}, false
}

func (c *TenantConfiguration) DefaultSensitiveLoggerValues() []string {
	return []string{
		c.UserConfig.APIKey,
		c.UserConfig.MasterKey,
	}
}

func (c *TenantConfiguration) Validate() error {
	if c.AppConfig.DatabaseURL == "" {
		return errors.New("DATABASE_URL is not set")
	}
	if c.AppName == "" {
		return errors.New("APP_NAME is not set")
	}
	if c.UserConfig.APIKey == "" {
		return errors.New("API_KEY is not set")
	}
	if c.UserConfig.MasterKey == "" {
		return errors.New("MASTER_KEY is not set")
	}
	if c.UserConfig.APIKey == c.UserConfig.MasterKey {
		return errors.New("MASTER_KEY cannot be the same as API_KEY")
	}
	return name.ValidateAppName(c.AppName)
}

func (c *TenantConfiguration) AfterUnmarshal() {
	if c.UserConfig.TokenStore.Secret == "" {
		c.UserConfig.TokenStore.Secret = c.UserConfig.MasterKey
	}

	// Propagate AppName
	if c.UserConfig.ForgotPassword.AppName == "" {
		c.UserConfig.ForgotPassword.AppName = c.AppName
	}

	// Propagate URLPrefix
	if c.UserConfig.ForgotPassword.URLPrefix == "" {
		c.UserConfig.ForgotPassword.URLPrefix = c.UserConfig.URLPrefix
	}
	if c.UserConfig.WelcomeEmail.URLPrefix == "" {
		c.UserConfig.WelcomeEmail.URLPrefix = c.UserConfig.URLPrefix
	}
	if c.UserConfig.SSO.URLPrefix == "" {
		c.UserConfig.SSO.URLPrefix = c.UserConfig.URLPrefix
	}
	if c.UserConfig.UserVerification.URLPrefix == "" {
		c.UserConfig.UserVerification.URLPrefix = c.UserConfig.URLPrefix
	}

	// Remove trailing slash in URLs
	c.UserConfig.URLPrefix = removeTrailingSlash(c.UserConfig.URLPrefix)
	c.UserConfig.ForgotPassword.URLPrefix = removeTrailingSlash(c.UserConfig.ForgotPassword.URLPrefix)
	c.UserConfig.WelcomeEmail.URLPrefix = removeTrailingSlash(c.UserConfig.WelcomeEmail.URLPrefix)
	c.UserConfig.UserVerification.URLPrefix = removeTrailingSlash(c.UserConfig.UserVerification.URLPrefix)
	c.UserConfig.SSO.URLPrefix = removeTrailingSlash(c.UserConfig.SSO.URLPrefix)
}

func GetTenantConfig(r *http.Request) TenantConfiguration {
	// TODO: Use msgpack instead of JSON
	s := r.Header.Get(coreHttp.HeaderTenantConfig)
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	var config TenantConfiguration
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		panic(err)
	}
	return config
}

func SetTenantConfig(r *http.Request, config *TenantConfiguration) {
	// TODO: Use msgpack instead of JSON
	bytes, err := json.Marshal(*config)
	if err != nil {
		panic(err)
	}
	r.Header.Set(coreHttp.HeaderTenantConfig, base64.StdEncoding.EncodeToString(bytes))
}

func DelTenantConfig(r *http.Request) {
	r.Header.Del(coreHttp.HeaderTenantConfig)
}

func removeTrailingSlash(url string) string {
	if strings.HasSuffix(url, "/") {
		return url[:len(url)-1]
	}

	return url
}
