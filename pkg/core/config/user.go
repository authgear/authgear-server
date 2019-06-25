package config

import (
	"net/url"
)

// UserConfiguration represents user-editable configuration
type UserConfiguration struct {
	Version          string                         `json:"version" yaml:"version" msg:"version"`
	APIKey           string                         `json:"api_key" yaml:"api_key" msg:"api_key"`
	MasterKey        string                         `json:"master_key" yaml:"master_key" msg:"master_key"`
	URLPrefix        string                         `json:"url_prefix" yaml:"url_prefix" msg:"url_prefix"`
	CORS             CORSConfiguration              `json:"cors" yaml:"cors" msg:"cors"`
	Auth             NewAuthConfiguration           `json:"auth" yaml:"auth" msg:"auth"`
	TokenStore       NewTokenStoreConfiguration     `json:"token_store" yaml:"token_store" msg:"token_store"`
	UserAudit        NewUserAuditConfiguration      `json:"user_audit" yaml:"user_audit" msg:"user_audit"`
	ForgotPassword   NewForgotPasswordConfiguration `json:"forgot_password" yaml:"forgot_password" msg:"forgot_password"`
	WelcomeEmail     NewWelcomeEmailConfiguration   `json:"welcome_email" yaml:"welcome_email" msg:"welcome_email"`
	SSO              NewSSOConfiguration            `json:"sso" yaml:"sso" msg:"sso"`
	UserVerification UserVerificationConfiguration  `json:"user_verification" yaml:"user_verification" msg:"user_verification"`
}

// CORSConfiguration represents CORS configuration.
// Currently we only support configuring origin.
// We may allow to support other headers in the future.
// The interpretation of origin is done by this library
// https://github.com/iawaknahc/originmatcher
type CORSConfiguration struct {
	Origin string `json:"origin" yaml:"origin" msg:"origin"`
}

type NewAuthConfiguration struct {
	LoginIDKeys       []string `json:"login_id_keys" yaml:"login_id_keys" msg:"login_id_keys"`
	CustomTokenSecret string   `json:"custom_token_secret" yaml:"custom_token_secret" msg:"custom_token_secret"`
}

type NewTokenStoreConfiguration struct {
	Secret string `json:"secret" yaml:"secret" msg:"secret"`
	Expiry int64  `json:"expiry" yaml:"expiry" msg:"expiry"`
}

type NewUserAuditConfiguration struct {
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

type NewForgotPasswordConfiguration struct {
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

type NewWelcomeEmailConfiguration struct {
	Enabled     bool   `json:"enabled" yaml:"enabled" msg:"enabled"`
	URLPrefix   string `json:"url_prefix" yaml:"url_prefix" msg:"url_prefix"`
	SenderName  string `json:"sender_name" yaml:"sender_name" msg:"sender_name"`
	Sender      string `json:"sender" yaml:"sender" msg:"sender"`
	Subject     string `json:"subject" yaml:"subject" msg:"subject"`
	ReplyToName string `json:"reply_to_name" yaml:"reply_to_name" msg:"reply_to_name"`
	ReplyTo     string `json:"reply_to" yaml:"reply_to" msg:"reply_to"`
	TextURL     string `json:"text_url" yaml:"text_url" msg:"text_url"`
	HTMLURL     string `json:"html_url" yaml:"html_url" msg:"html_url"`
}

type NewSSOConfiguration struct {
	URLPrefix            string                     `json:"url_prefix" yaml:"url_prefix" msg:"url_prefix"`
	JSSDKCDNURL          string                     `json:"js_sdk_cdn_url" yaml:"js_sdk_cdn_url" msg:"js_sdk_cdn_url"`
	StateJWTSecret       string                     `json:"state_jwt_secret" yaml:"state_jwt_secret" msg:"state_jwt_secret"`
	AutoLinkProviderKeys []string                   `json:"auto_link_provider_keys" yaml:"auto_link_provider_keys" msg:"auto_link_provider_keys"`
	AllowedCallbackURLs  []string                   `json:"allowed_callback_urls" yaml:"allowed_callback_urls" msg:"allowed_callback_urls"`
	Providers            []SSOProviderConfiguration `json:"providers" yaml:"providers" msg:"providers"`
}

func (s *NewSSOConfiguration) APIEndpoint() string {
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

type SSOProviderConfiguration struct {
	Name         string `json:"name" yaml:"name" msg:"name"`
	ClientID     string `json:"client_id" yaml:"client_id" msg:"client_id"`
	ClientSecret string `json:"client_secret" yaml:"client_secret" msg:"client_secret"`
	Scope        string `json:"scope" yaml:"scope" msg:"scope"`
}

type UserVerificationConfiguration struct {
	URLPrefix        string                             `json:"url_prefix" yaml:"url_prefix" msg:"url_prefix"`
	AutoUpdate       bool                               `json:"auto_update" yaml:"auto_update" msg:"auto_update"`
	AutoSendOnSignup bool                               `json:"auto_send_on_signup" yaml:"auto_send_on_signup" msg:"auto_send_on_signup"`
	AutoSendOnUpdate bool                               `json:"auto_send_on_update" yaml:"auto_send_on_update" msg:"auto_send_on_update"`
	Required         bool                               `json:"required" yaml:"required" msg:"required"`
	Criteria         string                             `json:"criteria" yaml:"criteria" msg:"criteria"`
	ErrorRedirect    string                             `json:"error_redirect" yaml:"error_redirect" msg:"error_redirect"`
	ErrorHTMLURL     string                             `json:"error_html_url" yaml:"error_html_url" msg:"error_html_url"`
	Keys             []UserVerificationKeyConfiguration `json:"keys" yaml:"keys" msg:"keys"`
}

func (c *UserVerificationConfiguration) ConfigForKey(key string) (UserVerificationKeyConfiguration, bool) {
	for _, keyConfig := range c.Keys {
		if keyConfig.Key == key {
			return keyConfig, true
		}
	}
	return UserVerificationKeyConfiguration{}, false
}

type UserVerificationKeyConfiguration struct {
	Key             string                                `json:"key" yaml:"key" msg:"key"`
	CodeFormat      string                                `json:"code_format" yaml:"code_format" msg:"code_format"`
	Expiry          int64                                 `json:"expiry" yaml:"expiry" msg:"expiry"`
	SuccessRedirect string                                `json:"success_redirect" yaml:"success_redirect" msg:"success_redirect"`
	SuccessHTMLURL  string                                `json:"success_html_url" yaml:"success_html_url" msg:"success_html_url"`
	ErrorRedirect   string                                `json:"error_redirect" yaml:"error_redirect" msg:"error_redirect"`
	ErrorHTMLURL    string                                `json:"error_html_url" yaml:"error_html_url" msg:"error_html_url"`
	Provider        string                                `json:"provider" yaml:"provider" msg:"provider"`
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
