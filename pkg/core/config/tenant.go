package config

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gopkg.in/yaml.v2"

	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/marshal"
	"github.com/skygeario/skygear-server/pkg/core/phone"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

//go:generate msgp -tests=false
type TenantConfiguration struct {
	APIVersion       string                   `json:"api_version,omitempty" yaml:"api_version" msg:"api_version"`
	AppID            string                   `json:"app_id,omitempty" yaml:"app_id" msg:"app_id"`
	AppName          string                   `json:"app_name,omitempty" yaml:"app_name" msg:"app_name"`
	Hook             *HookTenantConfiguration `json:"hook,omitempty" yaml:"hook" msg:"hook" default_zero_value:"true"`
	DatabaseConfig   *DatabaseConfiguration   `json:"database_config,omitempty" yaml:"database_config" msg:"database_config" default_zero_value:"true"`
	AppConfig        *AppConfiguration        `json:"app_config,omitempty" yaml:"app_config" msg:"app_config" default_zero_value:"true"`
	TemplateItems    []TemplateItem           `json:"template_items,omitempty" yaml:"template_items" msg:"template_items"`
	Hooks            []Hook                   `json:"hooks,omitempty" yaml:"hooks" msg:"hooks"`
	DeploymentRoutes []DeploymentRoute        `json:"deployment_routes,omitempty" yaml:"deployment_routes" msg:"deployment_routes"`
}

type Hook struct {
	Event string `json:"event,omitempty" yaml:"event" msg:"event"`
	URL   string `json:"url,omitempty" yaml:"url" msg:"url"`
}

type DeploymentRoute struct {
	Version    string                 `json:"version,omitempty" yaml:"version" msg:"version"`
	Path       string                 `json:"path,omitempty" yaml:"path" msg:"path"`
	Type       string                 `json:"type,omitempty" yaml:"type" msg:"type"`
	TypeConfig map[string]interface{} `json:"type_config,omitempty" yaml:"type_config" msg:"type_config"`
}

type TemplateItemType string

type TemplateItem struct {
	Type        TemplateItemType `json:"type,omitempty" yaml:"type" msg:"type"`
	LanguageTag string           `json:"language_tag,omitempty" yaml:"language_tag" msg:"language_tag"`
	Key         string           `json:"key,omitempty" yaml:"key" msg:"key"`
	URI         string           `json:"uri,omitempty" yaml:"uri" msg:"uri"`
	Digest      string           `json:"digest" yaml:"digest" msg:"digest"`
}

func NewTenantConfigurationFromYAML(r io.Reader) (*TenantConfiguration, error) {
	decoder := yaml.NewDecoder(r)
	var j map[string]interface{}
	err := decoder.Decode(&j)
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return NewTenantConfigurationFromJSON(bytes.NewReader(b), false)
}

func NewTenantConfigurationFromJSON(r io.Reader, raw bool) (*TenantConfiguration, error) {
	if raw {
		decoder := json.NewDecoder(r)
		config := TenantConfiguration{}
		err := decoder.Decode(&config)
		if err != nil {
			return nil, err
		}
		return &config, nil
	}

	addDetails := func(err error) error {
		causes := validation.ErrorCauses(err)
		msgs := make([]string, len(causes))
		for i, c := range causes {
			msgs[i] = fmt.Sprintf("%s: %s", c.Pointer, c.Message)
		}
		err = errors.WithDetails(
			err,
			errors.Details{"validation_error": errors.SafeDetail.Value(msgs)},
		)
		return err
	}

	config, err := ParseTenantConfiguration(r)
	if err != nil {
		return nil, addDetails(err)
	}

	config.AfterUnmarshal()

	err = config.PostValidate()
	if err != nil {
		return nil, addDetails(err)
	}

	return config, nil
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
	// The Scan implemented by TenantConfiguration always call AfterUnmarshal.
	config, err := NewTenantConfigurationFromJSON(bytes.NewReader(b), false)
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
	for _, provider := range c.AppConfig.Identity.OAuth.Providers {
		if provider.ID == id {
			return provider, true
		}
	}
	return OAuthProviderConfiguration{}, false
}

func (c *TenantConfiguration) DefaultSensitiveLoggerValues() []string {
	values := make([]string, len(c.AppConfig.Clients)+1)
	values[0] = c.AppConfig.MasterKey
	i := 1
	for _, clientConfig := range c.AppConfig.Clients {
		values[i] = clientConfig.ClientID()
		i++
	}

	values = append(values,
		c.AppConfig.Authentication.Secret,
		c.AppConfig.Identity.OAuth.StateJWTSecret,
		c.AppConfig.Hook.Secret,
		c.DatabaseConfig.DatabaseURL,
		c.DatabaseConfig.DatabaseSchema,
		c.AppConfig.SMTP.Host,
		c.AppConfig.SMTP.Login,
		c.AppConfig.SMTP.Password,
		c.AppConfig.Twilio.AccountSID,
		c.AppConfig.Twilio.AuthToken,
		c.AppConfig.Nexmo.APIKey,
		c.AppConfig.Nexmo.APISecret,
	)
	oauthSecrets := make([]string, len(c.AppConfig.Identity.OAuth.Providers)*2)
	for i, oauthConfig := range c.AppConfig.Identity.OAuth.Providers {
		oauthSecrets[i*2] = oauthConfig.ClientID
		oauthSecrets[i*2+1] = oauthConfig.ClientSecret
	}
	values = append(values, oauthSecrets...)
	return values
}

// nolint: gocyclo
func (c *TenantConfiguration) PostValidate() error {
	fail := func(kind validation.ErrorCauseKind, msg string, pointerTokens ...interface{}) error {
		return validation.NewValidationFailed("invalid tenant config", []validation.ErrorCause{{
			Kind:    kind,
			Pointer: validation.JSONPointer(pointerTokens...),
			Message: msg,
		}})
	}

	// Validate complex AppConfiguration
	for key, clientConfig := range c.AppConfig.Clients {
		if clientConfig.ClientID() == c.AppConfig.MasterKey {
			return fail(validation.ErrorGeneral, "master key must not be same as client_id", "user_config", "master_key")
		}

		if clientConfig.RefreshTokenLifetime() < clientConfig.AccessTokenLifetime() {
			return fail(
				validation.ErrorGeneral,
				"refresh token lifetime must be greater than or equal to access token lifetime",
				"user_config", "clients", key, "refresh_token_lifetime")
		}
	}

	for _, verifyKeyConfig := range c.AppConfig.UserVerification.LoginIDKeys {
		ok := false
		for _, loginIDKey := range c.AppConfig.Identity.LoginID.Keys {
			if loginIDKey.Key == verifyKeyConfig.Key {
				ok = true
				break
			}
		}
		if !ok {
			return fail(
				validation.ErrorGeneral,
				"cannot verify disallowed login ID key",
				"user_config", "user_verification", "login_id_keys", verifyKeyConfig.Key)
		}
	}

	// Validate OAuth
	seenOAuthProviderID := map[string]struct{}{}
	for i, provider := range c.AppConfig.Identity.OAuth.Providers {
		// Ensure ID is not duplicate.
		if _, ok := seenOAuthProviderID[provider.ID]; ok {
			return fail(
				validation.ErrorGeneral,
				"duplicated OAuth provider",
				"user_config", "identity", "oauth", "providers", i)
		}
		seenOAuthProviderID[provider.ID] = struct{}{}
	}

	// Validate AuthenticationConfiguration
	seenAuthenticator := map[string]struct{}{}
	for i, a := range c.AppConfig.Authentication.PrimaryAuthenticators {
		if _, ok := seenAuthenticator[a]; ok {
			return fail(
				validation.ErrorGeneral,
				"duplicated authenticator",
				"user_config", "authentication", "primary_authenticators", i)
		}
		seenAuthenticator[a] = struct{}{}
	}
	for i, a := range c.AppConfig.Authentication.SecondaryAuthenticators {
		if _, ok := seenAuthenticator[a]; ok {
			return fail(
				validation.ErrorGeneral,
				"duplicated authenticator",
				"user_config", "authentication", "secondary_authenticators", i)
		}
		seenAuthenticator[a] = struct{}{}
	}

	// Validate AuthUICountryCallingCodeConfiguration
	countryCallingCodeDefaultOK := false
	for _, code := range c.AppConfig.AuthUI.CountryCallingCode.Values {
		if code == c.AppConfig.AuthUI.CountryCallingCode.Default {
			countryCallingCodeDefaultOK = true
		}
	}
	if !countryCallingCodeDefaultOK {
		return fail(
			validation.ErrorGeneral,
			"default country calling code is unlisted",
			"user_config", "auth_ui", "country_calling_code", "default",
		)
	}

	return nil
}

// nolint: gocyclo
// AfterUnmarshal should not be called before persisting the tenant config
// This function updates the tenant config with default value which provide
// features default behavior
func (c *TenantConfiguration) AfterUnmarshal() {

	marshal.UpdateNilFieldsWithZeroValue(c)

	// Set default dislay app name
	if c.AppConfig.DisplayAppName == "" {
		c.AppConfig.DisplayAppName = c.AppName
	}

	// Set default SessionConfiguration values
	if c.AppConfig.Session.Lifetime == 0 {
		c.AppConfig.Session.Lifetime = 86400
	}
	if c.AppConfig.Session.IdleTimeout == 0 {
		c.AppConfig.Session.IdleTimeout = 300
	}

	// Set default APIClientConfiguration values
	for i, clientConfig := range c.AppConfig.Clients {
		if clientConfig.AccessTokenLifetime() == 0 {
			clientConfig.SetAccessTokenLifetime(1800)
		}
		if clientConfig.RefreshTokenLifetime() == 0 {
			clientConfig.SetRefreshTokenLifetime(86400)
			if clientConfig.AccessTokenLifetime() > clientConfig.RefreshTokenLifetime() {
				clientConfig.SetRefreshTokenLifetime(clientConfig.AccessTokenLifetime())
			}
		}
		c.AppConfig.Clients[i] = clientConfig
	}

	// Set default AuthConfiguration
	if c.AppConfig.Identity.LoginID.Keys == nil {
		c.AppConfig.Identity.LoginID.Keys = []LoginIDKeyConfiguration{
			LoginIDKeyConfiguration{Key: "email", Type: LoginIDKeyType(metadata.Email)},
			LoginIDKeyConfiguration{Key: "phone", Type: LoginIDKeyType(metadata.Phone)},
			LoginIDKeyConfiguration{Key: "username", Type: LoginIDKeyType(metadata.Username)},
		}
	}

	if c.AppConfig.Identity.LoginID.Types.Email.CaseSensitive == nil {
		d := false
		c.AppConfig.Identity.LoginID.Types.Email.CaseSensitive = &d
	}
	if c.AppConfig.Identity.LoginID.Types.Email.BlockPlusSign == nil {
		d := false
		c.AppConfig.Identity.LoginID.Types.Email.BlockPlusSign = &d
	}
	if c.AppConfig.Identity.LoginID.Types.Email.IgnoreDotSign == nil {
		d := false
		c.AppConfig.Identity.LoginID.Types.Email.IgnoreDotSign = &d
	}

	if c.AppConfig.Identity.LoginID.Types.Username.BlockReservedUsernames == nil {
		d := true
		c.AppConfig.Identity.LoginID.Types.Username.BlockReservedUsernames = &d
	}
	if c.AppConfig.Identity.LoginID.Types.Username.ASCIIOnly == nil {
		d := true
		c.AppConfig.Identity.LoginID.Types.Username.ASCIIOnly = &d
	}
	if c.AppConfig.Identity.LoginID.Types.Username.CaseSensitive == nil {
		d := false
		c.AppConfig.Identity.LoginID.Types.Username.CaseSensitive = &d
	}

	// Set default minimum and maximum
	for i, config := range c.AppConfig.Identity.LoginID.Keys {
		if config.Maximum == nil {
			config.Maximum = new(int)
			*config.Maximum = 1
		}
		c.AppConfig.Identity.LoginID.Keys[i] = config
	}

	// Set default AuthenticationConfiguration
	if len(c.AppConfig.Authentication.Identities) == 0 {
		c.AppConfig.Authentication.Identities = []string{
			"oauth",
			"login_id",
		}
	}
	if len(c.AppConfig.Authentication.PrimaryAuthenticators) == 0 {
		c.AppConfig.Authentication.PrimaryAuthenticators = []string{
			"password",
		}
	}
	if c.AppConfig.Authentication.SecondaryAuthenticators == nil {
		c.AppConfig.Authentication.SecondaryAuthenticators = []string{
			"totp",
			"oob_otp",
			"bearer_token",
		}
	}
	if c.AppConfig.Authentication.SecondaryAuthenticationMode == "" {
		c.AppConfig.Authentication.SecondaryAuthenticationMode = SecondaryAuthenticationModeIfExists
	}

	// Set default AuthenticatorConfiguration
	if c.AppConfig.Authenticator.TOTP.Maximum == nil {
		c.AppConfig.Authenticator.TOTP.Maximum = new(int)
		*c.AppConfig.Authenticator.TOTP.Maximum = 99
	}
	if c.AppConfig.Authenticator.OOB.SMS.Maximum == nil {
		c.AppConfig.Authenticator.OOB.SMS.Maximum = new(int)
		*c.AppConfig.Authenticator.OOB.SMS.Maximum = 99
	}
	if c.AppConfig.Authenticator.OOB.Email.Maximum == nil {
		c.AppConfig.Authenticator.OOB.Email.Maximum = new(int)
		*c.AppConfig.Authenticator.OOB.Email.Maximum = 99
	}
	if c.AppConfig.Authenticator.BearerToken.ExpireInDays == 0 {
		c.AppConfig.Authenticator.BearerToken.ExpireInDays = 30
	}
	if c.AppConfig.Authenticator.RecoveryCode.Count == 0 {
		c.AppConfig.Authenticator.RecoveryCode.Count = 16
	}

	// Set default AuthenticatorOOBConfiguration
	emailMsg := c.AppConfig.Authenticator.OOB.Email.Message
	if emailMsg["subject"] == "" {
		emailMsg["subject"] = "Email Verification Instruction"
	}

	// Set default user verification settings
	if c.AppConfig.UserVerification.Criteria == "" {
		c.AppConfig.UserVerification.Criteria = UserVerificationCriteriaAny
	}
	for i, config := range c.AppConfig.UserVerification.LoginIDKeys {
		if config.CodeFormat == "" {
			config.CodeFormat = UserVerificationCodeFormatComplex
		}
		if config.Expiry == 0 {
			config.Expiry = 3600 // 1 hour
		}
		if config.EmailMessage["subject"] == "" {
			config.EmailMessage["subject"] = "Verification instruction"
		}
		c.AppConfig.UserVerification.LoginIDKeys[i] = config
	}

	// Set default WelcomeEmailConfiguration
	if c.AppConfig.WelcomeEmail.Destination == "" {
		c.AppConfig.WelcomeEmail.Destination = WelcomeEmailDestinationFirst
	}
	emailMsg = c.AppConfig.WelcomeEmail.Message
	if emailMsg["subject"] == "" {
		emailMsg["subject"] = "Welcome!"
	}

	// Set default ForgotPasswordConfiguration
	emailMsg = c.AppConfig.ForgotPassword.EmailMessage
	if emailMsg["subject"] == "" {
		emailMsg["subject"] = "Reset password instruction"
	}
	if c.AppConfig.ForgotPassword.ResetCodeLifetime == 0 {
		// https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html#step-3-send-a-token-over-a-side-channel
		// OWASP suggests the lifetime is no more than 20 minutes
		c.AppConfig.ForgotPassword.ResetCodeLifetime = 1200
	}

	// Set default SMTPConfiguration
	if c.AppConfig.SMTP.Mode == "" {
		c.AppConfig.SMTP.Mode = SMTPModeNormal
	}
	if c.AppConfig.SMTP.Port == 0 {
		c.AppConfig.SMTP.Port = 25
	}

	// Set default MessagesConfiguration
	emailMsg = c.AppConfig.Messages.Email
	if emailMsg["sender"] == "" {
		emailMsg["sender"] = "no-reply@skygear.io"
	}

	// Set type to id
	// Set default scope for OAuth Provider
	for i, provider := range c.AppConfig.Identity.OAuth.Providers {
		if provider.ID == "" {
			c.AppConfig.Identity.OAuth.Providers[i].ID = string(provider.Type)
		}
		switch provider.Type {
		case OAuthProviderTypeGoogle:
			if provider.Scope == "" {
				// https://developers.google.com/identity/protocols/googlescopes#google_sign-in
				c.AppConfig.Identity.OAuth.Providers[i].Scope = "openid profile email"
			}
		case OAuthProviderTypeFacebook:
			if provider.Scope == "" {
				// https://developers.facebook.com/docs/facebook-login/permissions/#reference-default
				// https://developers.facebook.com/docs/facebook-login/permissions/#reference-email
				c.AppConfig.Identity.OAuth.Providers[i].Scope = "email"
			}
		case OAuthProviderTypeInstagram:
			if provider.Scope == "" {
				// https://www.instagram.com/developer/authorization/
				c.AppConfig.Identity.OAuth.Providers[i].Scope = "basic"
			}
		case OAuthProviderTypeLinkedIn:
			if provider.Scope == "" {
				// https://docs.microsoft.com/en-us/linkedin/shared/integrations/people/profile-api?context=linkedin/compliance/context
				// https://docs.microsoft.com/en-us/linkedin/shared/integrations/people/primary-contact-api?context=linkedin/compliance/context
				c.AppConfig.Identity.OAuth.Providers[i].Scope = "r_liteprofile r_emailaddress"
			}
		case OAuthProviderTypeAzureADv2:
			if provider.Scope == "" {
				// https://docs.microsoft.com/en-us/azure/active-directory/develop/v2-permissions-and-consent#openid-connect-scopes
				c.AppConfig.Identity.OAuth.Providers[i].Scope = "openid profile email"
			}
		case OAuthProviderTypeApple:
			if provider.Scope == "" {
				c.AppConfig.Identity.OAuth.Providers[i].Scope = "email"
			}
		}
	}

	// Set default Auth UI configuration
	if c.AppConfig.AuthUI.CountryCallingCode.Values == nil {
		c.AppConfig.AuthUI.CountryCallingCode.Values = phone.CountryCallingCodes
	}
	if c.AppConfig.AuthUI.CountryCallingCode.Default == "" {
		c.AppConfig.AuthUI.CountryCallingCode.Default = c.AppConfig.AuthUI.CountryCallingCode.Values[0]
	}

	// Set default hook timeout
	if c.Hook.SyncHookTimeout == 0 {
		c.Hook.SyncHookTimeout = 5
	}
	if c.Hook.SyncHookTotalTimeout == 0 {
		c.Hook.SyncHookTotalTimeout = 10
	}
}

func ReadTenantConfig(r *http.Request) TenantConfiguration {
	s := r.Header.Get(coreHttp.HeaderTenantConfig)
	config, err := NewTenantConfigurationFromStdBase64Msgpack(s)
	if err != nil {
		panic(err)
	}
	return *config
}

func WriteTenantConfig(r *http.Request, config *TenantConfiguration) {
	if config == nil {
		r.Header.Del(coreHttp.HeaderTenantConfig)
	} else {
		value, err := config.StdBase64Msgpack()
		if err != nil {
			panic(err)
		}
		r.Header.Set(coreHttp.HeaderTenantConfig, value)
	}
}

// AppConfiguration represents user-editable configuration
type AppConfiguration struct {
	APIVersion       string                         `json:"api_version,omitempty" yaml:"api_version" msg:"api_version"`
	DisplayAppName   string                         `json:"display_app_name,omitempty" yaml:"display_app_name" msg:"display_app_name"`
	Clients          []OAuthClientConfiguration     `json:"clients,omitempty" yaml:"clients" msg:"clients"`
	MasterKey        string                         `json:"master_key,omitempty" yaml:"master_key" msg:"master_key"`
	Session          *SessionConfiguration          `json:"session,omitempty" yaml:"session" msg:"session" default_zero_value:"true"`
	CORS             *CORSConfiguration             `json:"cors,omitempty" yaml:"cors" msg:"cors" default_zero_value:"true"`
	AuthAPI          *AuthAPIConfiguration          `json:"auth_api,omitempty" yaml:"auth_api" msg:"auth_api" default_zero_value:"true"`
	Authentication   *AuthenticationConfiguration   `json:"authentication,omitempty" yaml:"authentication" msg:"authentication" default_zero_value:"true"`
	AuthUI           *AuthUIConfiguration           `json:"auth_ui,omitempty" yaml:"auth_ui" msg:"auth_ui" default_zero_value:"true"`
	OIDC             *OIDCConfiguration             `json:"oidc,omitempty" yaml:"oidc" msg:"oidc" default_zero_value:"true"`
	Authenticator    *AuthenticatorConfiguration    `json:"authenticator,omitempty" yaml:"authenticator" msg:"authenticator" default_zero_value:"true"`
	ForgotPassword   *ForgotPasswordConfiguration   `json:"forgot_password,omitempty" yaml:"forgot_password" msg:"forgot_password" default_zero_value:"true"`
	WelcomeEmail     *WelcomeEmailConfiguration     `json:"welcome_email,omitempty" yaml:"welcome_email" msg:"welcome_email" default_zero_value:"true"`
	Identity         *IdentityConfiguration         `json:"identity,omitempty" yaml:"identity" msg:"identity" default_zero_value:"true"`
	UserVerification *UserVerificationConfiguration `json:"user_verification,omitempty" yaml:"user_verification" msg:"user_verification" default_zero_value:"true"`
	Hook             *HookAppConfiguration          `json:"hook,omitempty" yaml:"hook" msg:"hook" default_zero_value:"true"`
	Messages         *MessagesConfiguration         `json:"messages,omitempty" yaml:"messages" msg:"messages" default_zero_value:"true"`
	SMTP             *SMTPConfiguration             `json:"smtp,omitempty" yaml:"smtp" msg:"smtp" default_zero_value:"true"`
	Twilio           *TwilioConfiguration           `json:"twilio,omitempty" yaml:"twilio" msg:"twilio" default_zero_value:"true"`
	Nexmo            *NexmoConfiguration            `json:"nexmo,omitempty" yaml:"nexmo" msg:"nexmo" default_zero_value:"true"`
	Localization     *LocalizationConfiguration     `json:"localization,omitempty" yaml:"localization" msg:"localization" default_zero_value:"true"`
	Asset            *AssetConfiguration            `json:"asset,omitempty" yaml:"asset" msg:"asset" default_zero_value:"true"`
}

type AssetConfiguration struct {
	Secret string `json:"secret,omitempty" yaml:"secret" msg:"secret"`
}

type OAuthClientConfiguration map[string]interface{}

func (c OAuthClientConfiguration) ClientID() string {
	if s, ok := c["client_id"].(string); ok {
		return s
	}
	return ""
}

func (c OAuthClientConfiguration) ClientURI() string {
	if s, ok := c["client_uri"].(string); ok {
		return s
	}
	return ""
}

func (c OAuthClientConfiguration) RedirectURIs() (out []string) {
	if arr, ok := c["redirect_uris"].([]interface{}); ok {
		for _, item := range arr {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
	}
	return
}

func (c OAuthClientConfiguration) AuthAPIUseCookie() bool {
	if b, ok := c["auth_api_use_cookie"].(bool); ok {
		return b
	}
	return false
}

func (c OAuthClientConfiguration) AccessTokenLifetime() int {
	if f64, ok := c["access_token_lifetime"].(float64); ok {
		return int(f64)
	}
	return 0
}

func (c OAuthClientConfiguration) SetAccessTokenLifetime(t int) {
	c["access_token_lifetime"] = float64(t)
}

func (c OAuthClientConfiguration) SetRefreshTokenLifetime(t int) {
	c["refresh_token_lifetime"] = float64(t)
}

func (c OAuthClientConfiguration) RefreshTokenLifetime() int {
	if f64, ok := c["refresh_token_lifetime"].(float64); ok {
		return int(f64)
	}
	return 0
}

func (c OAuthClientConfiguration) GrantTypes() (out []string) {
	if arr, ok := c["grant_types"].([]interface{}); ok {
		for _, item := range arr {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
	}
	return out
}

func (c OAuthClientConfiguration) ResponseTypes() (out []string) {
	if arr, ok := c["response_types"].([]interface{}); ok {
		for _, item := range arr {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
	}
	return out
}

func (c OAuthClientConfiguration) PostLogoutRedirectURIs() (out []string) {
	if arr, ok := c["post_logout_redirect_uris"].([]interface{}); ok {
		for _, item := range arr {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
	}
	return out
}

type SessionConfiguration struct {
	Lifetime            int     `json:"lifetime,omitempty" yaml:"lifetime" msg:"lifetime"`
	IdleTimeoutEnabled  bool    `json:"idle_timeout_enabled,omitempty" yaml:"idle_timeout_enabled" msg:"idle_timeout_enabled"`
	IdleTimeout         int     `json:"idle_timeout" yaml:"idle_timeout" msg:"idle_timeout"`
	CookieDomain        *string `json:"cookie_domain,omitempty" yaml:"cookie_domain" msg:"cookie_domain"`
	CookieNonPersistent bool    `json:"cookie_non_persistent,omitempty" yaml:"cookie_non_persistent" msg:"cookie_non_persistent"`
}

// CORSConfiguration represents CORS configuration.
// Currently we only support configuring origin.
// We may allow to support other headers in the future.
// The interpretation of origin is done by this library
// https://github.com/iawaknahc/originmatcher
type CORSConfiguration struct {
	Origin string `json:"origin,omitempty" yaml:"origin" msg:"origin"`
}

type OIDCConfiguration struct {
	Keys []OIDCSigningKeyConfiguration `json:"keys,omitempty" yaml:"keys" msg:"keys"`
}

type OIDCSigningKeyConfiguration struct {
	KID        string `json:"kid,omitempty" yaml:"kid" msg:"kid"`
	PublicKey  string `json:"public_key,omitempty" yaml:"public_key" msg:"public_key"`
	PrivateKey string `json:"private_key,omitempty" yaml:"private_key" msg:"private_key"`
}

type ForgotPasswordConfiguration struct {
	EmailMessage      EmailMessageConfiguration `json:"email_message,omitempty" yaml:"email_message" msg:"email_message" default_zero_value:"true"`
	SMSMessage        SMSMessageConfiguration   `json:"sms_message,omitempty" yaml:"sms_message" msg:"sms_message" default_zero_value:"true"`
	ResetCodeLifetime int                       `json:"reset_code_lifetime,omitempty" yaml:"reset_code_lifetime" msg:"reset_code_lifetime"`
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
	Enabled     bool                      `json:"enabled,omitempty" yaml:"enabled" msg:"enabled"`
	Message     EmailMessageConfiguration `json:"message,omitempty" yaml:"message" msg:"message" default_zero_value:"true"`
	Destination WelcomeEmailDestination   `json:"destination,omitempty" yaml:"destination" msg:"destination"`
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
	AutoSendOnSignup bool                               `json:"auto_send_on_signup,omitempty" yaml:"auto_send_on_signup" msg:"auto_send_on_signup"`
	Criteria         UserVerificationCriteria           `json:"criteria,omitempty" yaml:"criteria" msg:"criteria"`
	LoginIDKeys      []UserVerificationKeyConfiguration `json:"login_id_keys,omitempty" yaml:"login_id_keys" msg:"login_id_keys"`
}

type UserVerificationCodeFormat string

const (
	UserVerificationCodeFormatNumeric UserVerificationCodeFormat = "numeric"
	UserVerificationCodeFormatComplex UserVerificationCodeFormat = "complex"
)

type UserVerificationKeyConfiguration struct {
	Key             string                     `json:"key,omitempty" yaml:"key" msg:"key"`
	CodeFormat      UserVerificationCodeFormat `json:"code_format,omitempty" yaml:"code_format" msg:"code_format"`
	Expiry          int64                      `json:"expiry,omitempty" yaml:"expiry" msg:"expiry"`
	SuccessRedirect string                     `json:"success_redirect,omitempty" yaml:"success_redirect" msg:"success_redirect"`
	ErrorRedirect   string                     `json:"error_redirect,omitempty" yaml:"error_redirect" msg:"error_redirect"`
	SMSMessage      SMSMessageConfiguration    `json:"sms_message,omitempty" yaml:"sms_message" msg:"sms_message" default_zero_value:"true"`
	EmailMessage    EmailMessageConfiguration  `json:"email_message,omitempty" yaml:"email_message" msg:"email_message" default_zero_value:"true"`
}

func (format UserVerificationCodeFormat) IsValid() bool {
	return format == UserVerificationCodeFormatNumeric || format == UserVerificationCodeFormatComplex
}

func (c *UserVerificationConfiguration) GetLoginIDKey(key string) (*UserVerificationKeyConfiguration, bool) {
	for _, config := range c.LoginIDKeys {
		if config.Key == key {
			return &config, true
		}
	}

	return nil, false
}

type HookAppConfiguration struct {
	Secret string `json:"secret,omitempty" yaml:"secret" msg:"secret"`
}

// DatabaseConfiguration is database configuration.
type DatabaseConfiguration struct {
	DatabaseURL    string `json:"database_url,omitempty" yaml:"database_url" msg:"database_url"`
	DatabaseSchema string `json:"database_schema,omitempty" yaml:"database_schema" msg:"database_schema"`
}

type SMTPMode string

const (
	SMTPModeNormal SMTPMode = "normal"
	SMTPModeSSL    SMTPMode = "ssl"
)

type SMTPConfiguration struct {
	Host     string   `json:"host,omitempty" yaml:"host" msg:"host" envconfig:"HOST"`
	Port     int      `json:"port,omitempty" yaml:"port" msg:"port" envconfig:"PORT"`
	Mode     SMTPMode `json:"mode,omitempty" yaml:"mode" msg:"mode" envconfig:"MODE"`
	Login    string   `json:"login,omitempty" yaml:"login" msg:"login" envconfig:"LOGIN"`
	Password string   `json:"password,omitempty" yaml:"password" msg:"password" envconfig:"PASSWORD"`
}

func (c SMTPConfiguration) IsValid() bool {
	return c.Host != ""
}

type TwilioConfiguration struct {
	AccountSID string `json:"account_sid,omitempty" yaml:"account_sid" msg:"account_sid" envconfig:"ACCOUNT_SID"`
	AuthToken  string `json:"auth_token,omitempty" yaml:"auth_token" msg:"auth_token" envconfig:"AUTH_TOKEN"`
}

func (c TwilioConfiguration) IsValid() bool {
	return c.AccountSID != "" && c.AuthToken != ""
}

type NexmoConfiguration struct {
	APIKey    string `json:"api_key,omitempty" yaml:"api_key" msg:"api_key" envconfig:"API_KEY"`
	APISecret string `json:"api_secret,omitempty" yaml:"api_secret" msg:"api_secret" envconfig:"API_SECRET"`
}

func (c NexmoConfiguration) IsValid() bool {
	return c.APIKey != "" && c.APISecret != ""
}

type HookTenantConfiguration struct {
	SyncHookTimeout      int `json:"sync_hook_timeout_second,omitempty" yaml:"sync_hook_timeout_second" msg:"sync_hook_timeout_second"`
	SyncHookTotalTimeout int `json:"sync_hook_total_timeout_second,omitempty" yaml:"sync_hook_total_timeout_second" msg:"sync_hook_total_timeout_second"`
}

var (
	_ sql.Scanner   = &TenantConfiguration{}
	_ driver.Valuer = &TenantConfiguration{}
)
