package config

import (
	"errors"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func ReadAppConfigOptionsFromConsole() *config.GenerateAppConfigOptions {
	opts := &config.GenerateAppConfigOptions{}

	opts.AppID = promptString{
		Title:        "App ID",
		DefaultValue: "my-app",
	}.Prompt()

	opts.PublicOrigin = promptString{
		Title:        "HTTP origin of authgear",
		DefaultValue: "http://localhost:3000",
	}.Prompt()

	return opts
}

func ReadOAuthClientConfigsFromConsole() (*config.GenerateOAuthClientConfigOptions, error) {
	portalOrigin := promptString{
		Title:        "HTTP origin of portal",
		DefaultValue: "http://portal.localhost:8000",
	}.Prompt()

	u, err := url.Parse(portalOrigin)
	if err != nil {
		return nil, errors.New("invalid portal origin")
	}
	u.Path = "/oauth-redirect"
	redirectURI := u.String()

	u.Path = "/"
	postLogoutRedirectURI := u.String()

	return &config.GenerateOAuthClientConfigOptions{
		Name:                  "Portal",
		RedirectURI:           redirectURI,
		PostLogoutRedirectURI: postLogoutRedirectURI,
		ApplicationType:       config.OAuthClientApplicationTypeTraditionalWeb,
	}, nil
}

func ReadPhoneOTPMode() config.AuthenticatorPhoneOTPMode {
	p := promptString{
		Title:        `Phone OTP Mode (sms, whatsapp, whatsapp_sms)`,
		DefaultValue: "sms",
		Validate: func(value string) error {
			validChoices := []string{"sms", "whatsapp", "whatsapp_sms"}
			for _, choice := range validChoices {
				if value == choice {
					return nil
				}
			}
			return errors.New("must enter 'sms', 'whatsapp', or 'whatsapp_sms'")
		},
	}
	input := p.Prompt()

	switch input {
	case "sms":
		return config.AuthenticatorPhoneOTPModeSMSOnly
	case "whatsapp":
		return config.AuthenticatorPhoneOTPModeWhatsappOnly
	case "whatsapp_sms":
		return config.AuthenticatorPhoneOTPModeWhatsappSMS
	default:
		// This case should never be reached due to validation
		return config.AuthenticatorPhoneOTPModeSMSOnly
	}
}

func ReadSkipEmailVerification() bool {
	return promptBool{
		Title:        "Would you like to turn off email verification? (In case you don't have SMTP credentials in your initial setup)",
		DefaultValue: false,
	}.Prompt()
}

func ReadSecretConfigOptionsFromConsole() *config.GenerateSecretConfigOptions {
	opts := &config.GenerateSecretConfigOptions{}

	opts.DatabaseURL = promptURL{
		Title:        "Database URL",
		DefaultValue: "postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable",
	}.Prompt()

	opts.DatabaseSchema = promptString{
		Title:        "Database schema",
		DefaultValue: "public",
	}.Prompt()

	opts.AuditDatabaseURL = promptURL{
		Title:        "Audit Database URL",
		DefaultValue: "postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable",
	}.Prompt()

	opts.AuditDatabaseSchema = promptString{
		Title:        "Audit Database schema",
		DefaultValue: "public",
	}.Prompt()

	opts.ElasticsearchURL = promptString{
		Title:        "Elasticsearch URL",
		DefaultValue: "http://localhost:9200",
	}.Prompt()

	opts.RedisURL = promptURL{
		Title:        "Redis URL",
		DefaultValue: "redis://localhost",
	}.Prompt()

	opts.AnalyticRedisURL = promptURL{
		Title:        "Redis URL for analytic",
		DefaultValue: "redis://localhost/1",
	}.Prompt()

	return opts
}
