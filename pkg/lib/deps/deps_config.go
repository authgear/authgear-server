package deps

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var configDeps = wire.NewSet(
	wire.FieldsOf(new(*config.Config), "AppConfig", "SecretConfig"),
	wire.FieldsOf(new(*config.AppConfig),
		"ID",
		"HTTP",
		"Hook",
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
		"Verification",
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
	),
	ProvideDefaultTemplateLanguage,
	secretDeps,
)

func ProvideDefaultTemplateLanguage(c *config.Config) template.DefaultTemplateLanguage {
	return template.DefaultTemplateLanguage(*c.AppConfig.Localization.FallbackLanguage)
}

var secretDeps = wire.NewSet(
	ProvideDatabaseCredentials,
	ProvideRedisCredentials,
	ProvideAdminAPIAuthKeyMaterials,
	ProvideOAuthClientCredentials,
	ProvideSMTPServerCredentials,
	ProvideTwilioCredentials,
	ProvideNexmoCredentials,
	ProvideOAuthKeyMaterials,
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

func ProvideAdminAPIAuthKeyMaterials(c *config.SecretConfig) *config.AdminAPIAuthKey {
	s, _ := c.LookupData(config.AdminAPIAuthKeyKey).(*config.AdminAPIAuthKey)
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

func ProvideOAuthKeyMaterials(c *config.SecretConfig) *config.OAuthKeyMaterials {
	s, _ := c.LookupData(config.OAuthKeyMaterialsKey).(*config.OAuthKeyMaterials)
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
