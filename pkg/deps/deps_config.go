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
	return c.LookupData(config.DatabaseCredentialsKey).(*config.DatabaseCredentials)
}
func ProvideRedisCredentials(c *config.SecretConfig) *config.RedisCredentials {
	return c.LookupData(config.RedisCredentialsKey).(*config.RedisCredentials)
}
func ProvideOAuthClientCredentials(c *config.SecretConfig) *config.OAuthClientCredentials {
	return c.LookupData(config.OAuthClientCredentialsKey).(*config.OAuthClientCredentials)
}
func ProvideSMTPServerCredentials(c *config.SecretConfig) *config.SMTPServerCredentials {
	return c.LookupData(config.SMTPServerCredentialsKey).(*config.SMTPServerCredentials)
}
func ProvideTwilioCredentials(c *config.SecretConfig) *config.TwilioCredentials {
	return c.LookupData(config.TwilioCredentialsKey).(*config.TwilioCredentials)
}
func ProvideNexmoCredentials(c *config.SecretConfig) *config.NexmoCredentials {
	return c.LookupData(config.NexmoCredentialsKey).(*config.NexmoCredentials)
}
func ProvideJWTKeyMaterials(c *config.SecretConfig) *config.JWTKeyMaterials {
	return c.LookupData(config.JWTKeyMaterialsKey).(*config.JWTKeyMaterials)
}
func ProvideOIDCKeyMaterials(c *config.SecretConfig) *config.OIDCKeyMaterials {
	return c.LookupData(config.OIDCKeyMaterialsKey).(*config.OIDCKeyMaterials)
}
func ProvideCSRFKeyMaterials(c *config.SecretConfig) *config.CSRFKeyMaterials {
	return c.LookupData(config.CSRFKeyMaterialsKey).(*config.CSRFKeyMaterials)
}
func ProvideWebhookKeyMaterials(c *config.SecretConfig) *config.WebhookKeyMaterials {
	return c.LookupData(config.WebhookKeyMaterialsKey).(*config.WebhookKeyMaterials)
}
