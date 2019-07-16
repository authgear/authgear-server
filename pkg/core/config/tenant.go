package config

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"

	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/name"
)

//go:generate msgp -tests=false
type TenantConfiguration struct {
	Version          string            `json:"version" yaml:"version" msg:"version"`
	AppName          string            `json:"app_name" yaml:"app_name" msg:"app_name"`
	AppConfig        AppConfiguration  `json:"app_config" yaml:"app_config" msg:"app_config"`
	UserConfig       UserConfiguration `json:"user_config" yaml:"user_config" msg:"user_config"`
	Hooks            []Hook            `json:"hooks" yaml:"hooks" msg:"hooks"`
	DeploymentRoutes []DeploymentRoute `json:"deployment_routes" yaml:"deployment_routes" msg:"deployment_routes"`
}

type Hook struct {
	Async   bool   `json:"async" yaml:"async" msg:"async"`
	Event   string `json:"event" yaml:"event" msg:"event"`
	URL     string `json:"url" yaml:"url" msg:"url"`
	Timeout int    `json:"timeout" yaml:"timeout" msg:"timeout"`
}

type DeploymentRoute struct {
	Version    string                 `json:"version" yaml:"version" msg:"version"`
	Path       string                 `json:"path" yaml:"path" msg:"path"`
	Type       string                 `json:"type" yaml:"type" msg:"type"`
	TypeConfig map[string]interface{} `json:"type_config" yaml:"type_config" msg:"type_config"`
}

func defaultAppConfiguration() AppConfiguration {
	return AppConfiguration{
		DatabaseURL: "postgres://postgres:@localhost/postgres?sslmode=disable",
		SMTP: SMTPConfiguration{
			Port: 25,
			Mode: "normal",
		},
	}
}

func defaultUserConfiguration() UserConfiguration {
	return UserConfiguration{
		CORS: CORSConfiguration{
			Origin: "*",
		},
		Auth: AuthConfiguration{
			// Default to email and username
			LoginIDKeys: map[string]LoginIDKeyConfiguration{
				"username": LoginIDKeyConfiguration{Type: LoginIDKeyTypeRaw},
				"email":    LoginIDKeyConfiguration{Type: LoginIDKeyType(metadata.Email)},
				"phone":    LoginIDKeyConfiguration{Type: LoginIDKeyType(metadata.Phone)},
			},
			AllowedRealms: []string{"default"},
		},
		ForgotPassword: ForgotPasswordConfiguration{
			SecureMatch:      false,
			Sender:           "no-reply@skygeario.com",
			Subject:          "Reset password instruction",
			ResetURLLifetime: 43200,
		},
		WelcomeEmail: WelcomeEmailConfiguration{
			Enabled:     false,
			Sender:      "no-reply@skygeario.com",
			Subject:     "Welcome!",
			Destination: WelcomeEmailDestinationFirst,
		},
		SSO: SSOConfiguration{
			OAuth: OAuthConfiguration{
				JSSDKCDNURL: "https://code.skygear.io/js/skygear/latest/skygear.min.js",
			},
		},
	}
}

type FromScratchOptions struct {
	AppName     string `envconfig:"APP_NAME"`
	DatabaseURL string `envconfig:"DATABASE_URL"`
	APIKey      string `envconfig:"API_KEY"`
	MasterKey   string `envconfig:"MASTER_KEY"`
}

func NewTenantConfigurationFromScratch(options FromScratchOptions) (*TenantConfiguration, error) {
	c := TenantConfiguration{
		AppConfig:  defaultAppConfiguration(),
		UserConfig: defaultUserConfiguration(),
	}
	c.Version = "1"

	c.AppName = options.AppName
	c.AppConfig.DatabaseURL = options.DatabaseURL
	c.UserConfig.APIKey = options.APIKey
	c.UserConfig.MasterKey = options.MasterKey

	c.AfterUnmarshal()
	err := c.Validate()
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func loadTenantConfigurationFromYAML(r io.Reader) (*TenantConfiguration, error) {
	decoder := yaml.NewDecoder(r)
	config := TenantConfiguration{
		AppConfig:  defaultAppConfiguration(),
		UserConfig: defaultUserConfiguration(),
	}
	err := decoder.Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func NewTenantConfigurationFromYAML(r io.Reader) (*TenantConfiguration, error) {
	config, err := loadTenantConfigurationFromYAML(r)
	if err != nil {
		return nil, err
	}

	config.AfterUnmarshal()
	err = config.Validate()
	if err != nil {
		return nil, err
	}
	return config, nil
}

func NewTenantConfigurationFromEnv() (*TenantConfiguration, error) {
	options := FromScratchOptions{}
	err := envconfig.Process("", &options)
	if err != nil {
		return nil, err
	}
	return NewTenantConfigurationFromScratch(options)
}

func NewTenantConfigurationFromYAMLAndEnv(open func() (io.Reader, error)) (*TenantConfiguration, error) {
	options := FromScratchOptions{}
	err := envconfig.Process("", &options)
	if err != nil {
		return nil, err
	}

	r, err := open()
	if err != nil {
		// Load from env directly
		return NewTenantConfigurationFromScratch(options)
	}
	defer func() {
		if rc, ok := r.(io.Closer); ok {
			rc.Close()
		}
	}()

	c, err := loadTenantConfigurationFromYAML(r)
	if err != nil {
		return nil, err
	}

	// Allow override from env
	if options.AppName != "" {
		c.AppName = options.AppName
	}
	if options.DatabaseURL != "" {
		c.AppConfig.DatabaseURL = options.DatabaseURL
	}
	if options.APIKey != "" {
		c.UserConfig.APIKey = options.APIKey
	}
	if options.MasterKey != "" {
		c.UserConfig.MasterKey = options.MasterKey
	}

	c.AfterUnmarshal()
	err = c.Validate()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func NewTenantConfigurationFromJSON(r io.Reader) (*TenantConfiguration, error) {
	decoder := json.NewDecoder(r)
	config := TenantConfiguration{
		AppConfig:  defaultAppConfiguration(),
		UserConfig: defaultUserConfiguration(),
	}
	err := decoder.Decode(&config)
	if err != nil {
		return nil, err
	}
	config.AfterUnmarshal()
	err = config.Validate()
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func NewTenantConfigurationFromStdBase64Msgpack(s string) (*TenantConfiguration, error) {
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	var config TenantConfiguration
	_, err = config.UnmarshalMsg(bytes)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *TenantConfiguration) Value() (driver.Value, error) {
	bytes, err := json.Marshal(*c)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (c *TenantConfiguration) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("Cannot convert %T to TenantConfiguration", value)
	}
	config, err := NewTenantConfigurationFromJSON(bytes.NewReader(b))
	if err != nil {
		return err
	}
	*c = *config
	return nil
}

func (c *TenantConfiguration) StdBase64Msgpack() (string, error) {
	bytes, err := c.MarshalMsg(nil)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

func (c *TenantConfiguration) GetOAuthProviderByID(id string) (OAuthProviderConfiguration, bool) {
	for _, provider := range c.UserConfig.SSO.OAuth.Providers {
		if provider.ID == id {
			return provider, true
		}
	}
	return OAuthProviderConfiguration{}, false
}

func (c *TenantConfiguration) DefaultSensitiveLoggerValues() []string {
	return []string{
		c.UserConfig.APIKey,
		c.UserConfig.MasterKey,
	}
}

// nolint: gocyclo
func (c *TenantConfiguration) Validate() error {
	if c.Version != "1" {
		return errors.New("Only version 1 is supported")
	}
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

	if len(c.UserConfig.Auth.LoginIDKeys) == 0 {
		return errors.New("LoginIDKeys cannot be empty")
	}
	for _, loginIDKeyConfig := range c.UserConfig.Auth.LoginIDKeys {
		if !loginIDKeyConfig.Type.IsValid() {
			return errors.New("Invalid LoginIDKeys type: " + string(loginIDKeyConfig.Type))
		}
		if *loginIDKeyConfig.Minimum > *loginIDKeyConfig.Maximum || *loginIDKeyConfig.Maximum <= 0 {
			return errors.New("Invalid LoginIDKeys amount range: " + string(loginIDKeyConfig.Type))
		}
	}

	for key, verifyConfig := range c.UserConfig.UserVerification.LoginIDKeys {
		keyConfig, ok := c.UserConfig.Auth.LoginIDKeys[key]
		if !ok {
			return errors.New("Cannot verify disallowed login ID key: " + key)
		}
		if metadataKey, valid := keyConfig.Type.MetadataKey(); !valid || (metadataKey != metadata.Email && metadataKey != metadata.Phone) {
			return errors.New("Cannot verify login ID key with unknown type: " + key)
		}
		if !verifyConfig.CodeFormat.IsValid() {
			return errors.New("Invalid verify code format for login ID key: " + key)
		}
		if !verifyConfig.Provider.IsValid() {
			return errors.New("Invalid verify code provider for login ID key: " + key)
		}
	}

	if err := name.ValidateAppName(c.AppName); err != nil {
		return err
	}

	if !c.UserConfig.UserVerification.Criteria.IsValid() {
		return errors.New("Invalid user verification criteria")
	}

	if !c.UserConfig.WelcomeEmail.Destination.IsValid() {
		return errors.New("Invalid welcome email destination")
	}

	if !c.AppConfig.SMTP.Mode.IsValid() {
		return errors.New("Invalid SMTP mode")
	}

	// Validate CustomToken
	if c.UserConfig.SSO.CustomToken.Enabled {
		if c.UserConfig.SSO.CustomToken.Issuer == "" {
			return errors.New("Must set Custom Token Issuer")
		}
		if c.UserConfig.SSO.CustomToken.Secret == "" {
			return errors.New("Must set Custom Token Secret")
		}
	}

	// Validate OAuth
	seenOAuthProviderID := map[string]struct{}{}
	for _, provider := range c.UserConfig.SSO.OAuth.Providers {
		// Ensure ID is set
		if provider.ID == "" {
			return fmt.Errorf("Missing OAuth Provider ID")
		}

		// Ensure ID is not duplicate.
		if _, ok := seenOAuthProviderID[provider.ID]; ok {
			return fmt.Errorf("Duplicate OAuth Provider: %s", provider.ID)
		}
		seenOAuthProviderID[provider.ID] = struct{}{}

		switch provider.Type {
		case OAuthProviderTypeGoogle:
			break
		case OAuthProviderTypeFacebook:
			break
		case OAuthProviderTypeInstagram:
			break
		case OAuthProviderTypeLinkedIn:
			break
		case OAuthProviderTypeAzureADv2:
			// Ensure tenant is set
			if provider.Tenant == "" {
				return errors.New("Must set Azure Tenant")
			}
		default:
			// Ensure Type is recognized
			return fmt.Errorf("Unknown OAuth Provider: %s", provider.Type)
		}

		if provider.ClientID == "" {
			return fmt.Errorf("OAuth Provider %s: missing client id", provider.ID)
		}
		if provider.ClientSecret == "" {
			return fmt.Errorf("OAuth Provider %s: missing client secret", provider.ID)
		}
		if provider.Scope == "" {
			return fmt.Errorf("OAuth Provider %s: missing scope", provider.ID)
		}
	}
	oauthIsEffective := len(c.UserConfig.SSO.OAuth.Providers) > 0
	if oauthIsEffective {
		if len(c.UserConfig.SSO.OAuth.AllowedCallbackURLs) <= 0 {
			return fmt.Errorf("Must specify OAuth callback URLs")
		}
	}

	return nil
}

// nolint: gocyclo
func (c *TenantConfiguration) AfterUnmarshal() {
	// Default token secret to master key
	if c.UserConfig.TokenStore.Secret == "" {
		c.UserConfig.TokenStore.Secret = c.UserConfig.MasterKey
	}
	// Default oauth state secret to master key
	if c.UserConfig.SSO.OAuth.StateJWTSecret == "" {
		c.UserConfig.SSO.OAuth.StateJWTSecret = c.UserConfig.MasterKey
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
	if c.UserConfig.SSO.OAuth.URLPrefix == "" {
		c.UserConfig.SSO.OAuth.URLPrefix = c.UserConfig.URLPrefix
	}
	if c.UserConfig.UserVerification.URLPrefix == "" {
		c.UserConfig.UserVerification.URLPrefix = c.UserConfig.URLPrefix
	}

	// Remove trailing slash in URLs
	c.UserConfig.URLPrefix = removeTrailingSlash(c.UserConfig.URLPrefix)
	c.UserConfig.ForgotPassword.URLPrefix = removeTrailingSlash(c.UserConfig.ForgotPassword.URLPrefix)
	c.UserConfig.WelcomeEmail.URLPrefix = removeTrailingSlash(c.UserConfig.WelcomeEmail.URLPrefix)
	c.UserConfig.UserVerification.URLPrefix = removeTrailingSlash(c.UserConfig.UserVerification.URLPrefix)
	c.UserConfig.SSO.OAuth.URLPrefix = removeTrailingSlash(c.UserConfig.SSO.OAuth.URLPrefix)

	// Set default value for login ID keys config
	for key, config := range c.UserConfig.Auth.LoginIDKeys {
		if config.Minimum == nil {
			config.Minimum = new(int)
			*config.Minimum = 0
		}
		if config.Maximum == nil {
			config.Maximum = new(int)
			if *config.Minimum == 0 {
				*config.Maximum = 1
			} else {
				*config.Maximum = *config.Minimum
			}
		}
		c.UserConfig.Auth.LoginIDKeys[key] = config
	}

	// Set default user verification settings
	if c.UserConfig.UserVerification.Criteria == "" {
		c.UserConfig.UserVerification.Criteria = UserVerificationCriteriaAny
	}
	for key, config := range c.UserConfig.UserVerification.LoginIDKeys {
		if config.CodeFormat == "" {
			config.CodeFormat = UserVerificationCodeFormatComplex
		}
		if config.Expiry == 0 {
			config.Expiry = 3600 // 1 hour
		}
		if config.ProviderConfig.Sender == "" {
			config.ProviderConfig.Sender = "no-reply@skygeario.com"
		}
		if config.ProviderConfig.Subject == "" {
			config.ProviderConfig.Subject = "Verification instruction"
		}
		c.UserConfig.UserVerification.LoginIDKeys[key] = config
	}

	// Set default welcome email destination
	if c.UserConfig.WelcomeEmail.Destination == "" {
		c.UserConfig.WelcomeEmail.Destination = WelcomeEmailDestinationFirst
	}

	// Set default smtp mode
	if c.AppConfig.SMTP.Mode == "" {
		c.AppConfig.SMTP.Mode = SMTPModeNormal
	}

	// Set type to id
	// Set default scope for OAuth Provider
	for i, provider := range c.UserConfig.SSO.OAuth.Providers {
		if provider.ID == "" {
			c.UserConfig.SSO.OAuth.Providers[i].ID = string(provider.Type)
		}
		switch provider.Type {
		case OAuthProviderTypeGoogle:
			if provider.Scope == "" {
				// https://developers.google.com/identity/protocols/googlescopes#google_sign-in
				c.UserConfig.SSO.OAuth.Providers[i].Scope = "profile email"
			}
		case OAuthProviderTypeFacebook:
			if provider.Scope == "" {
				// https://developers.facebook.com/docs/facebook-login/permissions/#reference-default
				// https://developers.facebook.com/docs/facebook-login/permissions/#reference-email
				c.UserConfig.SSO.OAuth.Providers[i].Scope = "default email"
			}
		case OAuthProviderTypeInstagram:
			if provider.Scope == "" {
				// https://www.instagram.com/developer/authorization/
				c.UserConfig.SSO.OAuth.Providers[i].Scope = "basic"
			}
		case OAuthProviderTypeLinkedIn:
			if provider.Scope == "" {
				// https://docs.microsoft.com/en-us/linkedin/shared/integrations/people/profile-api?context=linkedin/compliance/context
				// https://docs.microsoft.com/en-us/linkedin/shared/integrations/people/primary-contact-api?context=linkedin/compliance/context
				c.UserConfig.SSO.OAuth.Providers[i].Scope = "r_liteprofile r_emailaddress"
			}
		case OAuthProviderTypeAzureADv2:
			if provider.Scope == "" {
				// https://docs.microsoft.com/en-us/azure/active-directory/develop/v2-permissions-and-consent#openid-connect-scopes
				c.UserConfig.SSO.OAuth.Providers[i].Scope = "profile email"
			}
		}
	}
}

func GetTenantConfig(r *http.Request) TenantConfiguration {
	s := r.Header.Get(coreHttp.HeaderTenantConfig)
	config, err := NewTenantConfigurationFromStdBase64Msgpack(s)
	if err != nil {
		panic(err)
	}
	return *config
}

func SetTenantConfig(r *http.Request, config *TenantConfiguration) {
	value, err := config.StdBase64Msgpack()
	if err != nil {
		panic(err)
	}
	r.Header.Set(coreHttp.HeaderTenantConfig, value)
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

// UserConfiguration represents user-editable configuration
type UserConfiguration struct {
	APIKey           string                        `json:"api_key" yaml:"api_key" msg:"api_key"`
	MasterKey        string                        `json:"master_key" yaml:"master_key" msg:"master_key"`
	URLPrefix        string                        `json:"url_prefix" yaml:"url_prefix" msg:"url_prefix"`
	CORS             CORSConfiguration             `json:"cors" yaml:"cors" msg:"cors"`
	Auth             AuthConfiguration             `json:"auth" yaml:"auth" msg:"auth"`
	TokenStore       TokenStoreConfiguration       `json:"token_store" yaml:"token_store" msg:"token_store"`
	UserAudit        UserAuditConfiguration        `json:"user_audit" yaml:"user_audit" msg:"user_audit"`
	ForgotPassword   ForgotPasswordConfiguration   `json:"forgot_password" yaml:"forgot_password" msg:"forgot_password"`
	WelcomeEmail     WelcomeEmailConfiguration     `json:"welcome_email" yaml:"welcome_email" msg:"welcome_email"`
	SSO              SSOConfiguration              `json:"sso" yaml:"sso" msg:"sso"`
	UserVerification UserVerificationConfiguration `json:"user_verification" yaml:"user_verification" msg:"user_verification"`
}

// CORSConfiguration represents CORS configuration.
// Currently we only support configuring origin.
// We may allow to support other headers in the future.
// The interpretation of origin is done by this library
// https://github.com/iawaknahc/originmatcher
type CORSConfiguration struct {
	Origin string `json:"origin" yaml:"origin" msg:"origin"`
}

type AuthConfiguration struct {
	LoginIDKeys   map[string]LoginIDKeyConfiguration `json:"login_id_keys" yaml:"login_id_keys" msg:"login_id_keys"`
	AllowedRealms []string                           `json:"allowed_realms" yaml:"allowed_realms" msg:"allowed_realms"`
}

type LoginIDKeyType string

const LoginIDKeyTypeRaw LoginIDKeyType = "raw"

func (t LoginIDKeyType) MetadataKey() (metadata.StandardKey, bool) {
	for _, key := range metadata.AllKeys() {
		if string(t) == string(key) {
			return key, true
		}
	}
	return "", false
}

func (t LoginIDKeyType) IsValid() bool {
	_, validKey := t.MetadataKey()
	return t == LoginIDKeyTypeRaw || validKey
}

type LoginIDKeyConfiguration struct {
	Type    LoginIDKeyType `json:"type" yaml:"type" msg:"type"`
	Minimum *int           `json:"minimum" yaml:"minimum" msg:"minimum"`
	Maximum *int           `json:"maximum" yaml:"maximum" msg:"maximum"`
}

type TokenStoreConfiguration struct {
	Secret string `json:"secret" yaml:"secret" msg:"secret"`
	Expiry int64  `json:"expiry" yaml:"expiry" msg:"expiry"`
}

type UserAuditConfiguration struct {
	Enabled         bool                  `json:"enabled" yaml:"enabled" msg:"enabled"`
	TrailHandlerURL string                `json:"trail_handler_url" yaml:"trail_handler_url" msg:"trail_handler_url"`
	Password        PasswordConfiguration `json:"password" yaml:"password" msg:"password"`
}

type PasswordConfiguration struct {
	MinLength             int      `json:"min_length" yaml:"min_length" msg:"min_length"`
	UppercaseRequired     bool     `json:"uppercase_required" yaml:"uppercase_required" msg:"uppercase_required"`
	LowercaseRequired     bool     `json:"lowercase_required" yaml:"lowercase_required" msg:"lowercase_required"`
	DigitRequired         bool     `json:"digit_required" yaml:"digit_required" msg:"digit_required"`
	SymbolRequired        bool     `json:"symbol_required" yaml:"symbol_required" msg:"symbol_required"`
	MinimumGuessableLevel int      `json:"minimum_guessable_level" yaml:"minimum_guessable_level" msg:"minimum_guessable_level"`
	ExcludedKeywords      []string `json:"excluded_keywords" yaml:"excluded_keywords" msg:"excluded_keywords"`
	// Do not know how to support fields because we do not
	// have them now
	// ExcludedFields     []string `json:"excluded_fields" yaml:"excluded_fields" msg:"excluded_fields"`
	HistorySize int `json:"history_size" yaml:"history_size" msg:"history_size"`
	HistoryDays int `json:"history_days" yaml:"history_days" msg:"history_days"`
	ExpiryDays  int `json:"expiry_days" yaml:"expiry_days" msg:"expiry_days"`
}

type ForgotPasswordConfiguration struct {
	AppName             string `json:"app_name" yaml:"app_name" msg:"app_name"`
	URLPrefix           string `json:"url_prefix" yaml:"url_prefix" msg:"url_prefix"`
	SecureMatch         bool   `json:"secure_match" yaml:"secure_match" msg:"secure_match"`
	SenderName          string `json:"sender_name" yaml:"sender_name" msg:"sender_name"`
	Sender              string `json:"sender" yaml:"sender" msg:"sender"`
	Subject             string `json:"subject" yaml:"subject" msg:"subject"`
	ReplyToName         string `json:"reply_to_name" yaml:"reply_to_name" msg:"reply_to_name"`
	ReplyTo             string `json:"reply_to" yaml:"reply_to" msg:"reply_to"`
	ResetURLLifetime    int    `json:"reset_url_lifetime" yaml:"reset_url_lifetime" msg:"reset_url_lifetime"`
	SuccessRedirect     string `json:"success_redirect" yaml:"success_redirect" msg:"success_redirect"`
	ErrorRedirect       string `json:"error_redirect" yaml:"error_redirect" msg:"error_redirect"`
	EmailTextURL        string `json:"email_text_url" yaml:"email_text_url" msg:"email_text_url"`
	EmailHTMLURL        string `json:"email_html_url" yaml:"email_html_url" msg:"email_html_url"`
	ResetHTMLURL        string `json:"reset_html_url" yaml:"reset_html_url" msg:"reset_html_url"`
	ResetSuccessHTMLURL string `json:"reset_success_html_url" yaml:"reset_success_html_url" msg:"reset_success_html_url"`
	ResetErrorHTMLURL   string `json:"reset_error_html_url" yaml:"reset_error_html_url" msg:"reset_error_html_url"`
}

type WelcomeEmailDestination string

const (
	WelcomeEmailDestinationFirst WelcomeEmailDestination = "first"
	WelcomeEmailDestinationAll   WelcomeEmailDestination = "all"
)

func (destination WelcomeEmailDestination) IsValid() bool {
	return destination == WelcomeEmailDestinationFirst || destination == WelcomeEmailDestinationAll
}

type WelcomeEmailConfiguration struct {
	Enabled     bool                    `json:"enabled" yaml:"enabled" msg:"enabled"`
	URLPrefix   string                  `json:"url_prefix" yaml:"url_prefix" msg:"url_prefix"`
	SenderName  string                  `json:"sender_name" yaml:"sender_name" msg:"sender_name"`
	Sender      string                  `json:"sender" yaml:"sender" msg:"sender"`
	Subject     string                  `json:"subject" yaml:"subject" msg:"subject"`
	ReplyToName string                  `json:"reply_to_name" yaml:"reply_to_name" msg:"reply_to_name"`
	ReplyTo     string                  `json:"reply_to" yaml:"reply_to" msg:"reply_to"`
	TextURL     string                  `json:"text_url" yaml:"text_url" msg:"text_url"`
	HTMLURL     string                  `json:"html_url" yaml:"html_url" msg:"html_url"`
	Destination WelcomeEmailDestination `json:"destination" yaml:"destination" msg:"destination"`
}

type SSOConfiguration struct {
	CustomToken CustomTokenConfiguration `json:"custom_token" yaml:"custom_token" msg:"custom_token"`
	OAuth       OAuthConfiguration       `json:"oauth" yaml:"oauth" msg:"oauth"`
}

type CustomTokenConfiguration struct {
	Enabled                    bool   `json:"enabled" yaml:"enabled" msg:"enabled"`
	Issuer                     string `json:"issuer" yaml:"issuer" msg:"issuer"`
	Audience                   string `json:"audience" yaml:"audience" msg:"audience"`
	Secret                     string `json:"secret" yaml:"secret" msg:"secret"`
	OnUserDuplicateAllowMerge  bool   `json:"on_user_duplicate_allow_merge" yaml:"on_user_duplicate_allow_merge" msg:"on_user_duplicate_allow_merge"`
	OnUserDuplicateAllowCreate bool   `json:"on_user_duplicate_allow_create" yaml:"on_user_duplicate_allow_create" msg:"on_user_duplicate_allow_create"`
}

type OAuthConfiguration struct {
	URLPrefix                      string                       `json:"url_prefix" yaml:"url_prefix" msg:"url_prefix"`
	JSSDKCDNURL                    string                       `json:"js_sdk_cdn_url" yaml:"js_sdk_cdn_url" msg:"js_sdk_cdn_url"`
	StateJWTSecret                 string                       `json:"state_jwt_secret" yaml:"state_jwt_secret" msg:"state_jwt_secret"`
	AllowedCallbackURLs            []string                     `json:"allowed_callback_urls" yaml:"allowed_callback_urls" msg:"allowed_callback_urls"`
	ExternalAccessTokenFlowEnabled bool                         `json:"external_access_token_flow_enabled" yaml:"external_access_token_flow_enabled" msg:"external_access_token_flow_enabled"`
	OnUserDuplicateAllowMerge      bool                         `json:"on_user_duplicate_allow_merge" yaml:"on_user_duplicate_allow_merge" msg:"on_user_duplicate_allow_merge"`
	OnUserDuplicateAllowCreate     bool                         `json:"on_user_duplicate_allow_create" yaml:"on_user_duplicate_allow_create" msg:"on_user_duplicate_allow_create"`
	Providers                      []OAuthProviderConfiguration `json:"providers" yaml:"providers" msg:"providers"`
}

func (s *OAuthConfiguration) APIEndpoint() string {
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

type OAuthProviderType string

const (
	OAuthProviderTypeGoogle    OAuthProviderType = "google"
	OAuthProviderTypeFacebook  OAuthProviderType = "facebook"
	OAuthProviderTypeInstagram OAuthProviderType = "instagram"
	OAuthProviderTypeLinkedIn  OAuthProviderType = "linkedin"
	OAuthProviderTypeAzureADv2 OAuthProviderType = "azureadv2"
)

type OAuthProviderConfiguration struct {
	ID           string            `json:"id" yaml:"id" msg:"id"`
	Type         OAuthProviderType `json:"type" yaml:"type" msg:"type"`
	ClientID     string            `json:"client_id" yaml:"client_id" msg:"client_id"`
	ClientSecret string            `json:"client_secret" yaml:"client_secret" msg:"client_secret"`
	Scope        string            `json:"scope" yaml:"scope" msg:"scope"`
	// Type specific fields
	Tenant string `json:"tenant" yaml:"tenant" msg:"tenant"`
}

type UserVerificationCriteria string

const (
	// Some login ID need to verified belonging to the user is verified
	UserVerificationCriteriaAny UserVerificationCriteria = "any"
	// All login IDs need to verified belonging to the user is verified
	UserVerificationCriteriaAll UserVerificationCriteria = "all"
)

func (criteria UserVerificationCriteria) IsValid() bool {
	return criteria == UserVerificationCriteriaAny || criteria == UserVerificationCriteriaAll
}

type UserVerificationConfiguration struct {
	URLPrefix        string                                      `json:"url_prefix" yaml:"url_prefix" msg:"url_prefix"`
	AutoSendOnSignup bool                                        `json:"auto_send_on_signup" yaml:"auto_send_on_signup" msg:"auto_send_on_signup"`
	Criteria         UserVerificationCriteria                    `json:"criteria" yaml:"criteria" msg:"criteria"`
	ErrorRedirect    string                                      `json:"error_redirect" yaml:"error_redirect" msg:"error_redirect"`
	ErrorHTMLURL     string                                      `json:"error_html_url" yaml:"error_html_url" msg:"error_html_url"`
	LoginIDKeys      map[string]UserVerificationKeyConfiguration `json:"login_id_keys" yaml:"login_id_keys" msg:"login_id_keys"`
}

type UserVerificationCodeFormat string

const (
	UserVerificationCodeFormatNumeric UserVerificationCodeFormat = "numeric"
	UserVerificationCodeFormatComplex UserVerificationCodeFormat = "complex"
)

func (format UserVerificationCodeFormat) IsValid() bool {
	return format == UserVerificationCodeFormatNumeric || format == UserVerificationCodeFormatComplex
}

type UserVerificationProvider string

const (
	UserVerificationProviderSMTP   UserVerificationProvider = "smtp"
	UserVerificationProviderTwilio UserVerificationProvider = "twilio"
	UserVerificationProviderNexmo  UserVerificationProvider = "nexmo"
)

func (format UserVerificationProvider) IsValid() bool {
	switch format {
	case UserVerificationProviderSMTP:
		return true
	case UserVerificationProviderTwilio:
		return true
	case UserVerificationProviderNexmo:
		return true
	}
	return false
}

type UserVerificationKeyConfiguration struct {
	CodeFormat      UserVerificationCodeFormat            `json:"code_format" yaml:"code_format" msg:"code_format"`
	Expiry          int64                                 `json:"expiry" yaml:"expiry" msg:"expiry"`
	SuccessRedirect string                                `json:"success_redirect" yaml:"success_redirect" msg:"success_redirect"`
	SuccessHTMLURL  string                                `json:"success_html_url" yaml:"success_html_url" msg:"success_html_url"`
	ErrorRedirect   string                                `json:"error_redirect" yaml:"error_redirect" msg:"error_redirect"`
	ErrorHTMLURL    string                                `json:"error_html_url" yaml:"error_html_url" msg:"error_html_url"`
	Provider        UserVerificationProvider              `json:"provider" yaml:"provider" msg:"provider"`
	ProviderConfig  UserVerificationProviderConfiguration `json:"provider_config" yaml:"provider_config" msg:"provider_config"`
}

type UserVerificationProviderConfiguration struct {
	Subject     string `json:"subject" yaml:"subject" msg:"subject"`
	Sender      string `json:"sender" yaml:"sender" msg:"sender"`
	SenderName  string `json:"sender_name" yaml:"sender_name" msg:"sender_name"`
	ReplyTo     string `json:"reply_to" yaml:"reply_to" msg:"reply_to"`
	ReplyToName string `json:"reply_to_name" yaml:"reply_to_name" msg:"reply_to_name"`
	TextURL     string `json:"text_url" yaml:"text_url" msg:"text_url"`
	HTMLURL     string `json:"html_url" yaml:"html_url" msg:"html_url"`
}

// AppConfiguration is configuration kept secret from the developer.
type AppConfiguration struct {
	DatabaseURL string              `json:"database_url" yaml:"database_url" msg:"database_url"`
	SMTP        SMTPConfiguration   `json:"smtp" yaml:"smtp" msg:"smtp"`
	Twilio      TwilioConfiguration `json:"twilio" yaml:"twilio" msg:"twilio"`
	Nexmo       NexmoConfiguration  `json:"nexmo" yaml:"nexmo" msg:"nexmo"`
}

type SMTPMode string

const (
	SMTPModeNormal SMTPMode = "normal"
	SMTPModeSSL    SMTPMode = "ssl"
)

func (mode SMTPMode) IsValid() bool {
	switch mode {
	case SMTPModeNormal:
		return true
	case SMTPModeSSL:
		return true
	}
	return false
}

type SMTPConfiguration struct {
	Host     string   `json:"host" yaml:"host" msg:"host"`
	Port     int      `json:"port" yaml:"port" msg:"port"`
	Mode     SMTPMode `json:"mode" yaml:"mode" msg:"mode"`
	Login    string   `json:"login" yaml:"login" msg:"login"`
	Password string   `json:"password" yaml:"password" msg:"password"`
}

type TwilioConfiguration struct {
	AccountSID string `json:"account_sid" yaml:"account_sid" msg:"account_sid"`
	AuthToken  string `json:"auth_token" yaml:"auth_token" msg:"auth_token"`
	From       string `json:"from" yaml:"from" msg:"from"`
}

type NexmoConfiguration struct {
	APIKey    string `json:"api_key" yaml:"api_key" msg:"api_key"`
	APISecret string `json:"secret" yaml:"secret" msg:"secret"`
	From      string `json:"from" yaml:"from" msg:"from"`
}

var (
	_ sql.Scanner   = &TenantConfiguration{}
	_ driver.Valuer = &TenantConfiguration{}
)
