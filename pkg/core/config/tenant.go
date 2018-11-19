package config

import (
	"encoding/base64"
	"net/http"

	"github.com/kelseyhightower/envconfig"
)

// TenantConfiguration is a mock struct of tenant configuration
//go:generate msgp -tests=false
type TenantConfiguration struct {
	DBConnectionStr string                    `msg:"DATABASE_URL" envconfig:"DATABASE_URL" json:"DATABASE_URL"`
	APIKey          string                    `msg:"API_KEY" envconfig:"API_KEY" json:"API_KEY"`
	MasterKey       string                    `msg:"MASTER_KEY" envconfig:"MASTER_KEY" json:"MASTER_KEY"`
	AppName         string                    `msg:"APP_NAME" envconfig:"APP_NAME" json:"APP_NAME"`
	CORSHost        string                    `msg:"CORS_HOST" envconfig:"CORS_HOST" json:"CORS_HOST" default:"*"`
	TokenStore      TokenStoreConfiguration   `json:"TOKEN_STORE" msg:"TOKEN_STORE"`
	UserProfile     UserProfileConfiguration  `json:"USER_PROFILE" msg:"USER_PROFILE"`
	UserAudit       UserAuditConfiguration    `json:"USER_AUDIT" msg:"USER_AUDIT"`
	SMTP            SMTPConfiguration         `json:"SMTP" msg:"SMTP"`
	WelcomeEmail    WelcomeEmailConfiguration `json:"WELCOME_EMAIL" msg:"WELCOME_EMAIL"`
	SSOSetting      SSOSetting                `json:"SSO_SETTING" msg:"SSO_SETTING"`
	SSOConfigs      []SSOConfiguration        `json:"SSO_CONFIGS" msg:"SSO_CONFIGS"`
}

type TokenStoreConfiguration struct {
	Secret string `msg:"SECRET" envconfig:"TOKEN_STORE_SECRET" json:"SECRET"`
	Expiry int64  `msg:"EXPIRY" envconfig:"TOKEN_STORE_EXPIRY" json:"EXPIRY"`
}

type UserProfileConfiguration struct {
	ImplName     string `msg:"IMPLEMENTATION" envconfig:"USER_PROFILE_IMPL_NAME" json:"IMPLEMENTATION"`
	ImplStoreURL string `msg:"IMPL_STORE_URL" envconfig:"USER_PROFILE_IMPL_STORE_URL" json:"IMPL_STORE_URL"`
}

type UserAuditConfiguration struct {
	Enabled         bool   `msg:"ENABLED" envconfig:"USER_AUDIT_ENABLED" json:"ENABLED"`
	TrailHandlerURL string `msg:"TRAIL_HANDLER_URL" envconfig:"USER_AUDIT_TRAIL_HANDLER_URL" json:"TRAIL_HANDLER_URL"`
}

type SMTPConfiguration struct {
	Host     string `msg:"HOST" envconfig:"SMTP_HOST" json:"HOST"`
	Port     int    `msg:"PORT" envconfig:"SMTP_PORT" json:"PORT" default:"25"`
	Mode     string `msg:"MODE" envconfig:"SMTP_MODE" json:"MODE" default:"normal"`
	Login    string `msg:"LOGIN" envconfig:"SMTP_LOGIN" json:"LOGIN"`
	Password string `msg:"PASSWORD" envconfig:"SMTP_PASSWORD" json:"PASSWORD"`
}

type WelcomeEmailConfiguration struct {
	Enabled     bool   `msg:"ENABLED" envconfig:"WELCOME_EMAIL_ENABLED" json:"ENABLED" default:"false"`
	SenderName  string `msg:"SENDER_NAME" envconfig:"WELCOME_EMAIL_SENDER_NAME" json:"SENDER_NAME"`
	Sender      string `msg:"SENDER" envconfig:"WELCOME_EMAIL_SENDER" json:"SENDER" default:"no-reply@skygeario.com"`
	Subject     string `msg:"SUBJECT" envconfig:"WELCOME_EMAIL_SUBJECT" json:"SUBJECT" default:"Welcome!"`
	ReplyToName string `msg:"REPLY_TO_NAME" envconfig:"WELCOME_EMAIL_REPLY_TO_NAME" json:"REPLY_TO_NAME"`
	ReplyTo     string `msg:"REPLY_TO" envconfig:"WELCOME_EMAIL_REPLY_TO" json:"REPLY_TO"`
	TextURL     string `msg:"TEXT_URL" envconfig:"WELCOME_EMAIL_TEXT_URL" json:"TEXT_URL"`
	HTMLURL     string `msg:"HTML_URL" envconfig:"WELCOME_EMAIL_HTML_URL" json:"HTML_URL"`
}

type SSOSetting struct {
	URLPrefix            string   `msg:"URL_PREFIX" envconfig:"SSO_SETTING_URL_PREFIX" json:"URL_PREFIX"`
	JSSDKCDNURL          string   `msg:"JS_SDK_CDN_URL" envconfig:"SSO_SETTING_JS_SDK_CDN_URL" json:"JS_SDK_CDN_URL"`
	StateJWTSecret       string   `msg:"STATE_JWT_SECRET" envconfig:"SSO_SETTING_STATE_JET_SECRET" json:"STATE_JWT_SECRET"`
	AutoLinkProviderKeys []string `msg:"AUTO_LINK_PROVIDER_KEYS" envconfig:"SSO_SETTING_AUTO_LINK_PROVIDER_KEYS" json:"AUTO_LINK_PROVIDER_KEYS"`
	AllowedCallbackURLs  []string `msg:"ALLOWED_CALLBACK_URLS" envconfig:"SSO_SETTING_ALLOWED_CALLBACK_URLS" json:"ALLOWED_CALLBACK_URLS"`
}

type SSOConfiguration struct {
	Name         string `msg:"NAME" ignored:"true" json:"NAME"`
	Enabled      bool   `msg:"ENABLED" envconfig:"ENABLED" json:"ENABLED"`
	ClientID     string `msg:"CLIENT_ID" envconfig:"CLIENT_ID" json:"CLIENT_ID"`
	ClientSecret string `msg:"CLIENT_SECRET" envconfig:"CLIENT_SECRET" json:"CLIENT_SECRET"`
	Scope        string `msg:"SCOPE" envconfig:"SCOPE" json:"SCOPE"`
}

func (c *TenantConfiguration) ReadFromEnv() error {
	return envconfig.Process("", c)
}

func (c *TenantConfiguration) DefaultSensitiveLoggerValues() []string {
	return []string{
		c.APIKey,
		c.MasterKey,
	}
}

func (c *TenantConfiguration) GetSSOConfigByName(name string) (config SSOConfiguration) {
	for _, SSOConfig := range c.SSOConfigs {
		if SSOConfig.Name == name {
			return SSOConfig
		}
	}
	return
}

func header(i interface{}) http.Header {
	switch i.(type) {
	case *http.Request:
		return (i.(*http.Request)).Header
	case http.ResponseWriter:
		return (i.(http.ResponseWriter)).Header()
	default:
		panic("Invalid type")
	}
}

func GetTenantConfig(i interface{}) TenantConfiguration {
	s := header(i).Get("X-Skygear-App-Config")
	var t TenantConfiguration
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}

	_, err = t.UnmarshalMsg(data)
	if err != nil {
		panic(err)
	}
	return t
}

func SetTenantConfig(i interface{}, t TenantConfiguration) {
	out, err := t.MarshalMsg(nil)
	if err != nil {
		panic(err)
	}
	header(i).Set("X-Skygear-App-Config", base64.StdEncoding.EncodeToString(out))
}

// NewTenantConfigurationFromEnv implements ConfigurationProvider
func NewTenantConfigurationFromEnv(_ *http.Request) (TenantConfiguration, error) {
	c := TenantConfiguration{}
	err := envconfig.Process("", &c)
	c.SSOConfigs = appendSSOConfigs()

	return c, err
}

func appendSSOConfigs() []SSOConfiguration {
	configs := make([]SSOConfiguration, 0)
	SSOProviderNames := []string{
		"google",
		"facebook",
		"instagram",
		"linkedin",
	}
	for _, name := range SSOProviderNames {
		config := SSOConfiguration{
			Name: name,
		}
		if err := envconfig.Process("sso_"+name, &config); err == nil {
			configs = append(configs, config)
		}
	}

	return configs
}
