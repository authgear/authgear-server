package config

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

// TenantConfiguration is a mock struct of tenant configuration
//go:generate msgp -tests=false
type TenantConfiguration struct {
	DBConnectionStr string                      `msg:"DATABASE_URL" envconfig:"DATABASE_URL" json:"DATABASE_URL"`
	APIKey          string                      `msg:"API_KEY" envconfig:"API_KEY" json:"API_KEY"`
	MasterKey       string                      `msg:"MASTER_KEY" envconfig:"MASTER_KEY" json:"MASTER_KEY"`
	AppName         string                      `msg:"APP_NAME" envconfig:"APP_NAME" json:"APP_NAME"`
	URLPrefix       string                      `msg:"URL_PREFIX" envconfig:"URL_PREFIX" json:"URL_PREFIX"`
	CORSHost        string                      `msg:"CORS_HOST" envconfig:"CORS_HOST" json:"CORS_HOST"`
	Auth            AuthConfiguration           `msg:"AUTH" json:"AUTH"`
	TokenStore      TokenStoreConfiguration     `json:"TOKEN_STORE" msg:"TOKEN_STORE"`
	UserAudit       UserAuditConfiguration      `json:"USER_AUDIT" msg:"USER_AUDIT"`
	SMTP            SMTPConfiguration           `json:"SMTP" msg:"SMTP"`
	Twilio          TwilioConfiguration         `json:"TWILIO" msg:"TWILIO"`
	Nexmo           NexmoConfiguration          `json:"NEXMO" msg:"NEXMO"`
	ForgotPassword  ForgotPasswordConfiguration `json:"FORGOT_PASSWORD" msg:"FORGOT_PASSWORD"`
	WelcomeEmail    WelcomeEmailConfiguration   `json:"WELCOME_EMAIL" msg:"WELCOME_EMAIL"`
	SSOSetting      SSOSetting                  `json:"SSO_SETTING" msg:"SSO_SETTING"`
	SSOProviders    []string                    `json:"SSO_PROVIDERS" envconfig:"SSO_PROVIDERS" msg:"SSO_PROVIDERS"`
	SSOConfigs      []SSOConfiguration          `json:"SSO_CONFIGS" msg:"SSO_CONFIGS"`
	UserVerify      UserVerifyConfiguration     `json:"USER_VERIFY" msg:"USER_VERIFY"`
}

type TokenStoreConfiguration struct {
	Secret string `msg:"SECRET" envconfig:"TOKEN_STORE_SECRET" json:"SECRET"`
	Expiry int64  `msg:"EXPIRY" envconfig:"TOKEN_STORE_EXPIRY" json:"EXPIRY"`
}

type AuthConfiguration struct {
	// RawLoginIDMetadataKeys is used when parsing from environment variables
	RawLoginIDMetadataKeys string `msg:"-" envconfig:"RAW_LOGIN_ID_METADATA_KEYS" json:"-"`
	// LoginIDMetadataKeys is used when unmarshal from json
	LoginIDMetadataKeys [][]string `msg:"LOGIN_ID_METADATA_KEYS" json:"LOGIN_ID_METADATA_KEYS"`
	CustomTokenSecret   string     `msg:"CUSTOM_TOKEN_SECRET" envconfig:"CUSTOM_TOKEN_SECRET" json:"CUSTOM_TOKEN_SECRET"`
}

type UserAuditConfiguration struct {
	Enabled             bool     `msg:"ENABLED" envconfig:"USER_AUDIT_ENABLED" json:"ENABLED"`
	TrailHandlerURL     string   `msg:"TRAIL_HANDLER_URL" envconfig:"USER_AUDIT_TRAIL_HANDLER_URL" json:"TRAIL_HANDLER_URL"`
	PwMinLength         int      `msg:"PW_MIN_LENGTH" envconfig:"USER_AUDIT_PW_MIN_LENGTH" json:"PW_MIN_LENGTH"`
	PwUppercaseRequired bool     `msg:"PW_UPPERCASE_REQUIRED" envconfig:"USER_AUDIT_PW_UPPERCASE_REQUIRED" json:"PW_UPPERCASE_REQUIRED"`
	PwLowercaseRequired bool     `msg:"PW_LOWERCASE_REQUIRED" envconfig:"USER_AUDIT_PW_LOWERCASE_REQUIRED" json:"PW_LOWERCASE_REQUIRED"`
	PwDigitRequired     bool     `msg:"PW_DIGIT_REQUIRED" envconfig:"USER_AUDIT_PW_DIGIT_REQUIRED" json:"PW_DIGIT_REQUIRED"`
	PwSymbolRequired    bool     `msg:"PW_SYMBOL_REQUIRED" envconfig:"USER_AUDIT_PW_SYMBOL_REQUIRED" json:"PW_SYMBOL_REQUIRED"`
	PwMinGuessableLevel int      `msg:"PW_MIN_GUESSABLE_LEVEL" envconfig:"USER_AUDIT_PW_MIN_GUESSABLE_LEVEL" json:"PW_MIN_GUESSABLE_LEVEL"`
	PwExcludedKeywords  []string `msg:"PW_EXCLUDED_KEYWORDS" envconfig:"USER_AUDIT_PW_EXCLUDED_KEYWORDS" json:"PW_EXCLUDED_KEYWORDS"`
	PwExcludedFields    []string `msg:"PW_EXCLUDED_FIELDS" envconfig:"USER_AUDIT_PW_EXCLUDED_FIELDS" json:"PW_EXCLUDED_FIELDS"`
	PwHistorySize       int      `msg:"PW_HISTORY_SIZE" envconfig:"USER_AUDIT_PW_HISTORY_SIZE" json:"PW_HISTORY_SIZE"`
	PwHistoryDays       int      `msg:"PW_HISTORY_DAYS" envconfig:"USER_AUDIT_PW_HISTORY_DAYS" json:"PW_HISTORY_DAYS"`
	PwExpiryDays        int      `msg:"PW_EXPIRY_DAYS" envconfig:"USER_AUDIT_PW_EXPIRY_DAYS" json:"PW_EXPIRY_DAYS"`
}

type SMTPConfiguration struct {
	Host     string `msg:"HOST" envconfig:"SMTP_HOST" json:"HOST"`
	Port     int    `msg:"PORT" envconfig:"SMTP_PORT" json:"PORT"`
	Mode     string `msg:"MODE" envconfig:"SMTP_MODE" json:"MODE"`
	Login    string `msg:"LOGIN" envconfig:"SMTP_LOGIN" json:"LOGIN"`
	Password string `msg:"PASSWORD" envconfig:"SMTP_PASSWORD" json:"PASSWORD"`
}

type ForgotPasswordConfiguration struct {
	AppName             string `msg:"APP_NAME" envconfig:"FORGOT_PASSWORD_APP_NAME" json:"APP_NAME"`
	URLPrefix           string `msg:"URL_PREFIX" envconfig:"FORGOT_PASSWORD_URL_PREFIX" json:"URL_PREFIX"`
	SecureMatch         bool   `msg:"SECURE_MATCH" envconfig:"FORGOT_PASSWORD_SECURE_MATCH" json:"SECURE_MATCH"`
	SenderName          string `msg:"SENDER_NAME" envconfig:"FORGOT_PASSWORD_SENDER_NAME" json:"SENDER_NAME"`
	Sender              string `msg:"SENDER" envconfig:"FORGOT_PASSWORD_SENDER" json:"SENDER"`
	Subject             string `msg:"SUBJECT" envconfig:"FORGOT_PASSWORD_SUBJECT" json:"SUBJECT"`
	ReplyToName         string `msg:"REPLY_TO_NAME" envconfig:"FORGOT_PASSWORD_REPLY_TO_NAME" json:"REPLY_TO_NAME"`
	ReplyTo             string `msg:"REPLY_TO" envconfig:"FORGOT_PASSWORD_REPLY_TO" json:"REPLY_TO"`
	ResetURLLifeTime    int    `msg:"RESET_URL_LIFE_TIME" envconfig:"FORGOT_PASSWORD_RESET_URL_LIFE_TIME" json:"RESET_URL_LIFE_TIME"`
	SuccessRedirect     string `msg:"SUCCESS_REDIRECT" envconfig:"FORGOT_PASSWORD_SUCCESS_REDIRECT" json:"SUCCESS_REDIRECT"`
	ErrorRedirect       string `msg:"ERROR_REDIRECT" envconfig:"FORGOT_PASSWORD_ERROR_REDIRECT" json:"ERROR_REDIRECT"`
	EmailTextURL        string `msg:"EMAIL_TEXT_URL" envconfig:"FORGOT_PASSWORD_EMAIL_TEXT_URL" json:"EMAIL_TEXT_URL"`
	EmailHTMLURL        string `msg:"EMAIL_HTML_URL" envconfig:"FORGOT_PASSWORD_EMAIL_HTML_URL" json:"EMAIL_HTML_URL"`
	ResetHTMLURL        string `msg:"RESET_HTML_URL" envconfig:"FORGOT_PASSWORD_RESET_HTML_URL" json:"RESET_HTML_URL"`
	ResetSuccessHTMLURL string `msg:"RESET_SUCCESS_HTML_URL" envconfig:"FORGOT_PASSWORD_RESET_SUCCESS_HTML_URL" json:"RESET_SUCCESS_HTML_URL"`
	ResetErrorHTMLURL   string `msg:"RESET_ERROR_HTML_URL" envconfig:"FORGOT_PASSWORD_RESET_ERROR_HTML_URL" json:"RESET_ERROR_HTML_URL"`
}

type WelcomeEmailConfiguration struct {
	Enabled     bool   `msg:"ENABLED" envconfig:"WELCOME_EMAIL_ENABLED" json:"ENABLED"`
	URLPrefix   string `msg:"URL_PREFIX" envconfig:"WELCOME_EMAIL_PRRFIX" json:"URL_PREFIX"`
	SenderName  string `msg:"SENDER_NAME" envconfig:"WELCOME_EMAIL_SENDER_NAME" json:"SENDER_NAME"`
	Sender      string `msg:"SENDER" envconfig:"WELCOME_EMAIL_SENDER" json:"SENDER"`
	Subject     string `msg:"SUBJECT" envconfig:"WELCOME_EMAIL_SUBJECT" json:"SUBJECT"`
	ReplyToName string `msg:"REPLY_TO_NAME" envconfig:"WELCOME_EMAIL_REPLY_TO_NAME" json:"REPLY_TO_NAME"`
	ReplyTo     string `msg:"REPLY_TO" envconfig:"WELCOME_EMAIL_REPLY_TO" json:"REPLY_TO"`
	TextURL     string `msg:"TEXT_URL" envconfig:"WELCOME_EMAIL_TEXT_URL" json:"TEXT_URL"`
	HTMLURL     string `msg:"HTML_URL" envconfig:"WELCOME_EMAIL_HTML_URL" json:"HTML_URL"`
}

type SSOSetting struct {
	URLPrefix            string   `msg:"URL_PREFIX" envconfig:"SSO_URL_PRRFIX" json:"URL_PREFIX"`
	JSSDKCDNURL          string   `msg:"JS_SDK_CDN_URL" envconfig:"SSO_JS_SDK_CDN_URL" json:"JS_SDK_CDN_URL"`
	StateJWTSecret       string   `msg:"STATE_JWT_SECRET" envconfig:"SSO_STATE_JWT_SECRET" json:"STATE_JWT_SECRET"`
	AutoLinkProviderKeys []string `msg:"AUTO_LINK_PROVIDER_KEYS" envconfig:"SSO_AUTO_LINK_PROVIDER_KEYS" json:"AUTO_LINK_PROVIDER_KEYS"`
	AllowedCallbackURLs  []string `msg:"ALLOWED_CALLBACK_URLS" envconfig:"SSO_ALLOWED_CALLBACK_URLS" json:"ALLOWED_CALLBACK_URLS"`
}

type SSOConfiguration struct {
	Name         string `msg:"NAME" ignored:"true" json:"NAME"`
	ClientID     string `msg:"CLIENT_ID" envconfig:"CLIENT_ID" json:"CLIENT_ID"`
	ClientSecret string `msg:"CLIENT_SECRET" envconfig:"CLIENT_SECRET" json:"CLIENT_SECRET"`
	Scope        string `msg:"SCOPE" envconfig:"SCOPE" json:"SCOPE"`
}

type UserVerifyConfiguration struct {
	URLPrefix        string                       `msg:"URL_PREFIX" envconfig:"VERIFY_URL_PREFIX" json:"URL_PREFIX"`
	AutoUpdate       bool                         `msg:"AUTO_UPDATE" envconfig:"VERIFY_AUTO_UPDATE" json:"AUTO_UPDATE"`
	AutoSendOnSignup bool                         `msg:"AUTO_SEND_SIGNUP" envconfig:"VERIFY_AUTO_SEND_SIGNUP" json:"AUTO_SEND_SIGNUP"`
	AutoSendOnUpdate bool                         `msg:"AUTO_SEND_UPDATE" envconfig:"VERIFY_AUTO_SEND_UPDATE" json:"AUTO_SEND_UPDATE"`
	Required         bool                         `msg:"REQUIRED" envconfig:"VERIFY_REQUIRED" json:"REQUIRED"`
	Criteria         string                       `msg:"CRITERIA" envconfig:"VERIFY_CRITERIA" json:"CRITERIA"`
	ErrorRedirect    string                       `msg:"ERROR_REDIRECT" envconfig:"VERIFY_ERROR_REDIRECT" json:"ERROR_REDIRECT"`
	ErrorHTMLURL     string                       `msg:"ERROR_HTML_URL" envconfig:"VERIFY_ERROR_HTML_URL" json:"ERROR_HTML_URL"`
	Keys             []string                     `msg:"KEYS" envconfig:"VERIFY_KEYS" json:"KEYS"`
	KeyConfigs       []UserVerifyKeyConfiguration `msg:"KEY_CONFIGS" json:"KEY_CONFIGS"`
}

func (u *UserVerifyConfiguration) ConfigForKey(key string) (UserVerifyKeyConfiguration, bool) {
	for _, c := range u.KeyConfigs {
		if c.Key == key {
			return c, true
		}
	}

	return UserVerifyKeyConfiguration{}, false
}

type UserVerifyKeyConfiguration struct {
	Key             string `msg:"KEY" ignored:"true" json:"KEY"`
	CodeFormat      string `msg:"CODE_FORMAT" envconfig:"CODE_FORMAT" json:"CODE_FORMAT"`
	Expiry          int64  `msg:"EXPIRY" envconfig:"EXPIRY" json:"EXPIRY"`
	SuccessRedirect string `msg:"SUCCESS_REDIRECT" envconfig:"SUCCESS_REDIRECT" json:"SUCCESS_REDIRECT"`
	SuccessHTMLURL  string `msg:"SUCCESS_HTML_URL" envconfig:"SUCCESS_HTML_URL" json:"SUCCESS_HTML_URL"`
	ErrorRedirect   string `msg:"ERROR_REDIRECT" envconfig:"ERROR_REDIRECT" json:"ERROR_REDIRECT"`
	ErrorHTMLURL    string `msg:"ERROR_HTML_URL" envconfig:"ERROR_HTML_URL" json:"ERROR_HTML_URL"`
	Provider        string `msg:"PROVIDER" envconfig:"PROVIDER" json:"PROVIDER"`

	// provider config
	ProviderConfig UserVerifyKeyProviderConfiguration `msg:"PROVIDER_CONFIG" json:"PROVIDER_CONFIG"`
}

type UserVerifyKeyProviderConfiguration struct {
	Subject     string `msg:"SUBJECT" envconfig:"SUBJECT" json:"SUBJECT"`
	Sender      string `msg:"SENDER" envconfig:"SENDER" json:"SENDER"`
	SenderName  string `msg:"SENDER_NAME" envconfig:"SENDER_NAME" json:"SENDER_NAME"`
	ReplyTo     string `msg:"REPLY_TO" envconfig:"REPLY_TO" json:"REPLY_TO"`
	ReplyToName string `msg:"REPLY_TO_NAME" envconfig:"REPLY_TO_NAME" json:"REPLY_TO_NAME"`
	TextURL     string `msg:"TEXT_URL" envconfig:"TEXT_URL" json:"TEXT_URL"`
	HTMLURL     string `msg:"HTML_URL" envconfig:"HTML_URL" json:"HTML_URL"`
}

type TwilioConfiguration struct {
	AccountSID string `msg:"ACCOUNT_SID" envconfig:"TWILIO_ACCOUNT_SID" json:"ACCOUNT_SID"`
	AuthToken  string `msg:"AUTH_TOKEN" envconfig:"TWILIO_AUTH_TOKEN" json:"AUTH_TOKEN"`
	From       string `msg:"FROM" envconfig:"TWILIO_FROM" json:"FROM"`
}

type NexmoConfiguration struct {
	APIKey    string `msg:"API_KEY" envconfig:"NEXMO_API_KEY" json:"API_KEY"`
	APISecret string `msg:"API_SECRET" envconfig:"NEXMO_API_SECRET" json:"API_SECRET"`
	From      string `msg:"FROM" envconfig:"NEXMO_FROM" json:"FROM"`
}

func NewTenantConfiguration() TenantConfiguration {
	return TenantConfiguration{
		DBConnectionStr: "postgres://postgres:@localhost/postgres?sslmode=disable",
		CORSHost:        "*",
		Auth: AuthConfiguration{
			LoginIDMetadataKeys: [][]string{[]string{"email"}, []string{"username"}},
		},
		SMTP: SMTPConfiguration{
			Port: 25,
			Mode: "normal",
		},
		ForgotPassword: ForgotPasswordConfiguration{
			SecureMatch:      false,
			Sender:           "no-reply@skygeario.com",
			Subject:          "Reset password instruction",
			ResetURLLifeTime: 43200,
		},
		WelcomeEmail: WelcomeEmailConfiguration{
			Enabled: false,
			Sender:  "no-reply@skygeario.com",
			Subject: "Welcome!",
		},
		SSOSetting: SSOSetting{
			JSSDKCDNURL: "https://code.skygear.io/js/skygear/latest/skygear.min.js",
		},
	}
}

func (c *TenantConfiguration) Validate() error {
	if c.DBConnectionStr == "" {
		return errors.New("DATABASE_URL is not set")
	}
	if c.AppName == "" {
		return errors.New("APP_NAME is not set")
	}
	if c.APIKey == "" {
		return errors.New("API_KEY is not set")
	}
	if c.MasterKey == "" {
		return errors.New("MASTER_KEY is not set")
	}
	if c.APIKey == c.MasterKey {
		return errors.New("MASTER_KEY cannot be the same as API_KEY")
	}
	if !regexp.MustCompile("^[A-Za-z0-9_]+$").MatchString(c.AppName) {
		return fmt.Errorf("APP_NAME '%s' contains invalid characters other than alphanumerics or underscores", c.AppName)
	}
	return nil
}

func (c *TenantConfiguration) AfterUnmarshal() {
	if c.TokenStore.Secret == "" {
		c.TokenStore.Secret = c.MasterKey
	}

	if c.ForgotPassword.AppName == "" {
		c.ForgotPassword.AppName = c.AppName
	}

	if c.ForgotPassword.URLPrefix == "" {
		c.ForgotPassword.URLPrefix = c.URLPrefix
	}

	if c.WelcomeEmail.URLPrefix == "" {
		c.WelcomeEmail.URLPrefix = c.URLPrefix
	}

	if c.UserVerify.URLPrefix == "" {
		c.UserVerify.URLPrefix = c.URLPrefix
	}

	if c.SSOSetting.URLPrefix == "" {
		c.SSOSetting.URLPrefix = c.URLPrefix
	}

	c.URLPrefix = c.sanitzeURL(c.URLPrefix)
	c.ForgotPassword.URLPrefix = c.sanitzeURL(c.ForgotPassword.URLPrefix)
	c.WelcomeEmail.URLPrefix = c.sanitzeURL(c.WelcomeEmail.URLPrefix)
	c.UserVerify.URLPrefix = c.sanitzeURL(c.UserVerify.URLPrefix)
	c.SSOSetting.URLPrefix = c.sanitzeURL(c.SSOSetting.URLPrefix)
}

func (c *TenantConfiguration) sanitzeURL(url string) string {
	if strings.HasSuffix(url, "/") {
		return url[:len(url)-1]
	}

	return url
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

func (c *TenantConfiguration) UnmarshalJSON(b []byte) error {
	type configAlias TenantConfiguration
	if err := json.Unmarshal(b, (*configAlias)(c)); err != nil {
		return err
	}
	c.AfterUnmarshal()
	err := c.Validate()
	return err
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

func GetUserVerifyKeyConfigFromEnv(key string) (config UserVerifyKeyConfiguration, err error) {
	config.Key = key
	prefix := "verify_keys_" + key
	if err = envconfig.Process(prefix, &config); err != nil {
		return
	}

	prefix = "verify_keys_" + key + "_provider"
	if err = envconfig.Process(prefix, &config.ProviderConfig); err != nil {
		return
	}

	return
}

// NewTenantConfigurationFromEnv implements ConfigurationProvider
func NewTenantConfigurationFromEnv(_ *http.Request) (c TenantConfiguration, err error) {
	c = NewTenantConfiguration()
	err = envconfig.Process("", &c)
	if err != nil {
		return
	}
	err = parseAuthRecordKeys(&c.Auth)
	if err != nil {
		return
	}
	getSSOSetting(&c.SSOSetting)
	getSSOConfigs(c.SSOProviders, &c.SSOConfigs)

	// Read user verify config
	for _, userVerifyKey := range c.UserVerify.Keys {
		var keyConfig UserVerifyKeyConfiguration
		if keyConfig, err = GetUserVerifyKeyConfigFromEnv(userVerifyKey); err != nil {
			return
		}

		c.UserVerify.KeyConfigs = append(c.UserVerify.KeyConfigs, keyConfig)
	}

	c.AfterUnmarshal()
	err = c.Validate()

	return
}

func getSSOSetting(ssoSetting *SSOSetting) {
	envconfig.Process("", ssoSetting)
	return
}

func (s SSOSetting) APIEndpoint() string {
	// URLPrefix can't be seen as skygear endpoint.
	// Consider URLPrefix = http://localhost:3001/auth
	// and skygear SDK use is as base endpint URL (in iframe_html and auth_handler_html).
	// And then, SDK may generate wrong action path base on this wrong endpoint (http://localhost:3001/auth).
	// So, this function will remote path part of URLPrefix
	u, err := url.Parse(s.URLPrefix)
	if err != nil {
		return s.URLPrefix
	}
	u.Path = ""
	return u.String()
}

func getSSOConfigs(providers []string, ssoConfigs *[]SSOConfiguration) {
	configs := make([]SSOConfiguration, 0)
	for _, name := range providers {
		config := SSOConfiguration{
			Name: name,
		}
		if err := envconfig.Process("sso_"+name, &config); err == nil {
			configs = append(configs, config)
		}
	}
	*ssoConfigs = configs
	return
}

func parseAuthRecordKeys(authConfiguration *AuthConfiguration) error {
	if authConfiguration.RawLoginIDMetadataKeys == "" {
		// use default setting
		return nil
	}

	splits := strings.Split(authConfiguration.RawLoginIDMetadataKeys, ",")
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
			return errors.New("Unexpected char in " + content)
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

	authConfiguration.LoginIDMetadataKeys = results

	return nil
}
