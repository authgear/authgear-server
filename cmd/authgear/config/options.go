package config

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/cliutil"
)

var Prompt_AppID = cliutil.PromptString{
	Title:                       "App ID",
	InteractiveDefaultUserInput: "my-app",
	NonInteractiveFlagName:      "app-id",
}

var Prompt_PublicOrigin = cliutil.PromptURL{
	Title:                       "HTTP origin of authgear",
	InteractiveDefaultUserInput: "http://localhost:3000",
	NonInteractiveFlagName:      "public-origin",
}

var Prompt_PortalOrigin = cliutil.PromptURL{
	Title:                       "HTTP origin of portal",
	InteractiveDefaultUserInput: "http://portal.localhost:8000",
	NonInteractiveFlagName:      "portal-origin",
}

var Prompt_PortalClientID = cliutil.PromptString{
	Title:                       "The client ID of portal. If left empty, generate a random one.",
	InteractiveDefaultUserInput: "",
	NonInteractiveFlagName:      "portal-client-id",
}

var Prompt_SiteadminClientID = cliutil.PromptString{
	Title:                       "The client ID of siteadmin. If left empty, do not generate the siteadmin client.",
	InteractiveDefaultUserInput: "",
	NonInteractiveFlagName:      "siteadmin-client-id",
}

var Prompt_SiteadminRedirectURI = cliutil.PromptOptionalURL{
	Title:                  "Siteadmin redirect URI",
	NonInteractiveFlagName: "siteadmin-redirect-uri",
}

var Prompt_SiteadminPostLogoutRedirectURI = cliutil.PromptOptionalURL{
	Title:                  "Siteadmin post logout redirect URI",
	NonInteractiveFlagName: "siteadmin-post-logout-redirect-uri",
}

var Prompt_PhoneOTPMode = cliutil.PromptString{
	Title:                       `Phone OTP Mode (sms, whatsapp, whatsapp_sms)`,
	InteractiveDefaultUserInput: "sms",
	NonInteractiveFlagName:      "phone-otp-mode",
	Validate: func(ctx context.Context, value string) error {
		validChoices := []string{"sms", "whatsapp", "whatsapp_sms"}
		for _, choice := range validChoices {
			if value == choice {
				return nil
			}
		}
		return errors.New("must enter 'sms', 'whatsapp', or 'whatsapp_sms'")
	},
}

var Prompt_DisableEmailVerification = cliutil.PromptBool{
	Title:                       "Would you like to turn off email verification? (In case you don't have SMTP credentials in your initial setup)",
	InteractiveDefaultUserInput: false,
	NonInteractiveFlagName:      "disable-email-verification",
}

var Prompt_DisablePublicSignup = cliutil.PromptBool{
	Title:                       "Would you like to turn off public signup? (If turned off, you have to provision the users yourself in the portal)",
	InteractiveDefaultUserInput: false,
	NonInteractiveFlagName:      "disable-public-signup",
}

var Prompt_SMTPHost = cliutil.PromptString{
	Title:                       "SMTP host",
	InteractiveDefaultUserInput: "",
	NonInteractiveFlagName:      "smtp-host",
}

var Prompt_SMTPPort = cliutil.PromptOptionalPort{
	Title:                  "SMTP port. e.g. 25, 587",
	NonInteractiveFlagName: "smtp-port",
}

var Prompt_SMTPUsername = cliutil.PromptString{
	Title:                       "SMTP username",
	InteractiveDefaultUserInput: "",
	NonInteractiveFlagName:      "smtp-username",
}

var Prompt_SMTPPassword = cliutil.PromptString{
	Title:                       "SMTP password",
	InteractiveDefaultUserInput: "",
	NonInteractiveFlagName:      "smtp-password",
}

var Prompt_SMTPSenderAddress = cliutil.PromptOptionalEmailAddress{
	Title:                  "SMTP sender address",
	NonInteractiveFlagName: "smtp-sender-address",
}

var Prompt_SearchImplementation = cliutil.PromptString{
	Title:                       "Select a service for searching (elasticsearch, postgresql, none)",
	InteractiveDefaultUserInput: string(config.SearchImplementationElasticsearch),
	NonInteractiveFlagName:      "search-implementation",
	Validate: func(ctx context.Context, value string) error {
		validChoices := []string{
			string(config.SearchImplementationElasticsearch),
			string(config.SearchImplementationPostgresql),
			string(config.SearchImplementationNone),
		}
		for _, choice := range validChoices {
			if value == choice {
				return nil
			}
		}
		return errors.New("must enter 'elasticsearch', 'postgresql' or 'none'")
	},
}

var Prompt_DatabaseURL = cliutil.PromptURL{
	Title: "Database URL",
	// #nosec G101 -- Local development example DSN shown in interactive prompts.
	InteractiveDefaultUserInput: "postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable",
	NonInteractiveFlagName:      "database-url",
}

var Prompt_DatabaseSchema = cliutil.PromptString{
	Title:                       "Database schema",
	InteractiveDefaultUserInput: "public",
	NonInteractiveFlagName:      "database-schema",
}

var Prompt_AuditDatabaseURL = cliutil.PromptURL{
	Title: "Audit Database URL",
	// #nosec G101 -- Local development example DSN shown in interactive prompts.
	InteractiveDefaultUserInput: "postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable",
	NonInteractiveFlagName:      "audit-database-url",
}

var Prompt_AuditDatabaseSchema = cliutil.PromptString{
	Title:                       "Audit Database schema",
	InteractiveDefaultUserInput: "public",
	NonInteractiveFlagName:      "audit-database-schema",
}

var Prompt_SearchDatabaseURL = cliutil.PromptURL{
	Title: "Search Database URL",
	// #nosec G101 -- Local development example DSN shown in interactive prompts.
	InteractiveDefaultUserInput: "postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable",
	NonInteractiveFlagName:      "search-database-url",
}

var Prompt_SearchDatabaseSchema = cliutil.PromptString{
	Title:                       "Search Database schema",
	InteractiveDefaultUserInput: "public",
	NonInteractiveFlagName:      "search-database-schema",
}

var Prompt_ElasticsearchURL = cliutil.PromptURL{
	Title:                       "Elasticsearch URL",
	InteractiveDefaultUserInput: "http://localhost:9200",
	NonInteractiveFlagName:      "elasticsearch-url",
}

var Prompt_RedisURL = cliutil.PromptURL{
	Title:                       "Redis URL",
	InteractiveDefaultUserInput: "redis://localhost",
	NonInteractiveFlagName:      "redis-url",
}

var Prompt_AnalyticRedisURL = cliutil.PromptURL{
	Title:                       "Redis URL for analytic",
	InteractiveDefaultUserInput: "redis://localhost/1",
	NonInteractiveFlagName:      "analytic-redis-url",
}

func ReadAppConfigOptionsFromConsole(ctx context.Context, cmd *cobra.Command) (*config.GenerateAppConfigOptions, error) {
	opts := &config.GenerateAppConfigOptions{}
	var err error

	opts.AppID, err = Prompt_AppID.Prompt(ctx, cmd)
	if err != nil {
		return nil, err
	}

	publicOrigin, err := Prompt_PublicOrigin.Prompt(ctx, cmd)
	if err != nil {
		return nil, err
	}
	opts.PublicOrigin = publicOrigin.String()

	return opts, nil
}

func ReadOAuthClientConfigsFromConsole(ctx context.Context, cmd *cobra.Command) ([]config.GenerateOAuthClientConfigOptions, error) {
	portalOrigin, err := Prompt_PortalOrigin.Prompt(ctx, cmd)
	if err != nil {
		return nil, err
	}

	portalRedirectURI, portalPostLogoutRedirectURI, err := makePortalOAuthEndpoints(portalOrigin)
	if err != nil {
		return nil, err
	}

	portalClientID, err := Prompt_PortalClientID.Prompt(ctx, cmd)
	if err != nil {
		return nil, err
	}

	siteadminClientID, err := Prompt_SiteadminClientID.Prompt(ctx, cmd)
	if err != nil {
		return nil, err
	}

	siteadminRedirectURI, err := Prompt_SiteadminRedirectURI.Prompt(ctx, cmd)
	if err != nil {
		return nil, err
	}

	siteadminPostLogoutRedirectURI, err := Prompt_SiteadminPostLogoutRedirectURI.Prompt(ctx, cmd)
	if err != nil {
		return nil, err
	}

	var opts []config.GenerateOAuthClientConfigOptions

	opts = append(opts, config.GenerateOAuthClientConfigOptions{
		Name:                  "Portal",
		ClientID:              portalClientID,
		RedirectURI:           portalRedirectURI,
		PostLogoutRedirectURI: portalPostLogoutRedirectURI,
		ApplicationType:       config.OAuthClientApplicationTypeTraditionalWeb,
	})

	// Siteadmin is opt-in because many deployments do not need it.
	// Requiring an explicit client ID avoids seeding production configs with
	// localhost redirect URIs when siteadmin is not actually intended to be used.
	siteadminOpts, err := makeOAuthClientConfigOptions(
		"siteadmin",
		"Site Admin",
		siteadminClientID,
		siteadminRedirectURI,
		siteadminPostLogoutRedirectURI,
		config.OAuthClientApplicationTypeSPA,
	)
	if err != nil {
		return nil, err
	}
	if siteadminOpts != nil {
		opts = append(opts, *siteadminOpts)
	}

	return opts, nil
}

func makePortalOAuthEndpoints(origin *url.URL) (string, string, error) {
	if origin == nil {
		return "", "", fmt.Errorf("portal origin is required")
	}
	redirectURI := *origin
	redirectURI.Path = "/oauth-redirect"

	postLogoutRedirectURI := *origin
	postLogoutRedirectURI.Path = "/"

	return redirectURI.String(), postLogoutRedirectURI.String(), nil
}

func makeOAuthClientConfigOptions(clientKind string, displayName string, clientID string, redirectURI *url.URL, postLogoutRedirectURI *url.URL, applicationType config.OAuthClientApplicationType) (*config.GenerateOAuthClientConfigOptions, error) {
	if clientID == "" {
		return nil, nil
	}
	if redirectURI == nil {
		return nil, fmt.Errorf("%s redirect uri is required when %s-client-id is provided", clientKind, clientKind)
	}
	if postLogoutRedirectURI == nil {
		return nil, fmt.Errorf("%s post logout redirect uri is required when %s-client-id is provided", clientKind, clientKind)
	}
	return &config.GenerateOAuthClientConfigOptions{
		Name:                  displayName,
		ClientID:              clientID,
		RedirectURI:           redirectURI.String(),
		PostLogoutRedirectURI: postLogoutRedirectURI.String(),
		ApplicationType:       applicationType,
	}, nil
}

func ReadPhoneOTPMode(ctx context.Context, cmd *cobra.Command) (config.AuthenticatorPhoneOTPMode, error) {
	input, err := Prompt_PhoneOTPMode.Prompt(ctx, cmd)
	if err != nil {
		return "", err
	}

	switch input {
	case "sms":
		return config.AuthenticatorPhoneOTPModeSMSOnly, nil
	case "whatsapp":
		return config.AuthenticatorPhoneOTPModeWhatsappOnly, nil
	case "whatsapp_sms":
		return config.AuthenticatorPhoneOTPModeWhatsappSMS, nil
	default:
		// This case should never be reached due to validation
		return config.AuthenticatorPhoneOTPModeSMSOnly, nil
	}
}

func ReadSkipEmailVerification(ctx context.Context, cmd *cobra.Command) (bool, error) {
	b, err := Prompt_DisableEmailVerification.Prompt(ctx, cmd)
	if err != nil {
		return false, err
	}
	return b, nil
}

func ReadSkipPublicSignup(ctx context.Context, cmd *cobra.Command) (bool, error) {
	b, err := Prompt_DisablePublicSignup.Prompt(ctx, cmd)
	if err != nil {
		return false, err
	}
	return b, nil
}

func ReadSearchImplementation(ctx context.Context, cmd *cobra.Command) (config.SearchImplementation, error) {
	s, err := Prompt_SearchImplementation.Prompt(ctx, cmd)
	if err != nil {
		return "", err
	}

	return config.SearchImplementation(s), nil
}

func ReadSMTPConfig(ctx context.Context, cmd *cobra.Command) (*config.SMTPServerCredentials, error) {
	host, err := Prompt_SMTPHost.Prompt(ctx, cmd)
	if err != nil {
		return nil, err
	}

	port, err := Prompt_SMTPPort.Prompt(ctx, cmd)
	if err != nil {
		return nil, err
	}

	username, err := Prompt_SMTPUsername.Prompt(ctx, cmd)
	if err != nil {
		return nil, err
	}

	password, err := Prompt_SMTPPassword.Prompt(ctx, cmd)
	if err != nil {
		return nil, err
	}

	sender, err := Prompt_SMTPSenderAddress.Prompt(ctx, cmd)
	if err != nil {
		return nil, err
	}

	if host != "" && port != nil && username != "" && password != "" && sender != "" {
		return &config.SMTPServerCredentials{
			Host:     host,
			Port:     *port,
			Username: username,
			Password: password,
			Sender:   sender,
		}, nil
	}

	return nil, nil
}

func ReadSecretConfigOptionsFromConsole(ctx context.Context, cmd *cobra.Command, searchImpl config.SearchImplementation) (*config.GenerateSecretConfigOptions, error) {
	opts := &config.GenerateSecretConfigOptions{}

	databaseURL, err := Prompt_DatabaseURL.Prompt(ctx, cmd)
	if err != nil {
		return nil, err
	}
	opts.DatabaseURL = databaseURL.String()

	databaseSchema, err := Prompt_DatabaseSchema.Prompt(ctx, cmd)
	if err != nil {
		return nil, err
	}
	opts.DatabaseSchema = databaseSchema

	auditDatabaseURL, err := Prompt_AuditDatabaseURL.Prompt(ctx, cmd)
	if err != nil {
		return nil, err
	}
	opts.AuditDatabaseURL = auditDatabaseURL.String()

	auditDatabaseSchema, err := Prompt_AuditDatabaseSchema.Prompt(ctx, cmd)
	if err != nil {
		return nil, err
	}
	opts.AuditDatabaseSchema = auditDatabaseSchema

	switch searchImpl {
	case config.SearchImplementationNone:
		break
	case config.SearchImplementationPostgresql:
		searchDatabaseURL, err := Prompt_SearchDatabaseURL.Prompt(ctx, cmd)
		if err != nil {
			return nil, err
		}
		opts.SearchDatabaseURL = searchDatabaseURL.String()

		searchDatabaseSchema, err := Prompt_SearchDatabaseSchema.Prompt(ctx, cmd)
		if err != nil {
			return nil, err
		}
		opts.SearchDatabaseSchema = searchDatabaseSchema
	case config.SearchImplementationElasticsearch:
		fallthrough
	default:
		elasticsearchURL, err := Prompt_ElasticsearchURL.Prompt(ctx, cmd)
		if err != nil {
			return nil, err
		}

		opts.ElasticsearchURL = elasticsearchURL.String()
	}

	redisURL, err := Prompt_RedisURL.Prompt(ctx, cmd)
	if err != nil {
		return nil, err
	}
	opts.RedisURL = redisURL.String()

	analyticRedisURL, err := Prompt_AnalyticRedisURL.Prompt(ctx, cmd)
	if err != nil {
		return nil, err
	}
	opts.AnalyticRedisURL = analyticRedisURL.String()

	return opts, nil
}
