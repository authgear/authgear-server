package deps

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/config"
)

var configDeps = wire.NewSet(
	wire.FieldsOf(new(*config.Config), "AppConfig", "SecretConfig"),
	wire.FieldsOf(new(*config.AppConfig),
		"ID",
		"Metadata",
		"HTTP",
		"Hook",
		"Template",
		"UI",
		"Localization",
		"Messaging",
		"Authentication",
		"Session",
		"OAuth",
		"Identity",
		"Authenticator",
		"ForgotPassword",
		"WelcomeMessage",
	),
	wire.FieldsOf(new(*config.IdentityConfig),
		"LoginID",
		"OAuth",
		"OnConflict",
	),
	wire.FieldsOf(new(*config.AuthenticatorConfig),
		"Password",
		"TOTP",
		"OOB",
		"BearerToken",
		"RecoveryCode",
	),
	secretDeps,
)

var secretDeps = wire.NewSet(
	ProvideDatabaseCredentials,
	ProvideRedisCredentials,
	ProvideOAuthClientCredentials,
	ProvideSMTPServerCredentials,
	ProvideTwilioCredentials,
	ProvideNexmoCredentials,
	ProvideJWTKeyMaterials,
	ProvideOIDCKeyMaterials,
	ProvideCSRFKeyMaterials,
	ProvideWebhookKeyMaterials,
)

func ProvideDatabaseCredentials(c *config.SecretConfig) *config.DatabaseCredentials {
	s, _ := c.LookupData(config.DatabaseCredentialsKey).(*config.DatabaseCredentials)
	return s
}

func ProvideRedisCredentials(c *config.SecretConfig) *config.RedisCredentials {
	s, _ := c.LookupData(config.RedisCredentialsKey).(*config.RedisCredentials)
	return s
}

func ProvideOAuthClientCredentials(c *config.SecretConfig) *config.OAuthClientCredentials {
	s, _ := c.LookupData(config.OAuthClientCredentialsKey).(*config.OAuthClientCredentials)
	return s
}

func ProvideSMTPServerCredentials(c *config.SecretConfig) *config.SMTPServerCredentials {
	s, _ := c.LookupData(config.SMTPServerCredentialsKey).(*config.SMTPServerCredentials)
	return s
}

func ProvideTwilioCredentials(c *config.SecretConfig) *config.TwilioCredentials {
	s, _ := c.LookupData(config.TwilioCredentialsKey).(*config.TwilioCredentials)
	return s
}

func ProvideNexmoCredentials(c *config.SecretConfig) *config.NexmoCredentials {
	s, _ := c.LookupData(config.NexmoCredentialsKey).(*config.NexmoCredentials)
	return s
}

func ProvideJWTKeyMaterials(c *config.SecretConfig) *config.JWTKeyMaterials {
	s, _ := c.LookupData(config.JWTKeyMaterialsKey).(*config.JWTKeyMaterials)
	return s
}

func ProvideOIDCKeyMaterials(c *config.SecretConfig) *config.OIDCKeyMaterials {
	s, _ := c.LookupData(config.OIDCKeyMaterialsKey).(*config.OIDCKeyMaterials)
	return s
}

func ProvideCSRFKeyMaterials(c *config.SecretConfig) *config.CSRFKeyMaterials {
	s, _ := c.LookupData(config.CSRFKeyMaterialsKey).(*config.CSRFKeyMaterials)
	return s
}

func ProvideWebhookKeyMaterials(c *config.SecretConfig) *config.WebhookKeyMaterials {
	s, _ := c.LookupData(config.WebhookKeyMaterialsKey).(*config.WebhookKeyMaterials)
	return s
}
