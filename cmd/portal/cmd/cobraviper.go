package cmd

import (
	"github.com/authgear/authgear-server/pkg/util/cobraviper"
)

var cvbinder *cobraviper.Binder

func GetBinder() *cobraviper.Binder {
	if cvbinder == nil {
		cvbinder = cobraviper.NewBinder()
	}
	return cvbinder
}

var ArgDatabaseURL = &cobraviper.StringArgument{
	ArgumentName: "database-url",
	EnvName:      "DATABASE_URL",
	Usage:        "Database URL",
}

var ArgDatabaseSchema = &cobraviper.StringArgument{
	ArgumentName: "database-schema",
	EnvName:      "DATABASE_SCHEMA",
	Usage:        "Database schema",
}

var ArgAuditDatabaseURL = &cobraviper.StringArgument{
	ArgumentName: "audit-database-url",
	EnvName:      "AUDIT_DATABASE_URL",
	Usage:        "Audit Database URL",
}

var ArgAuditDatabaseSchema = &cobraviper.StringArgument{
	ArgumentName: "audit-database-schema",
	EnvName:      "AUDIT_DATABASE_SCHEMA",
	Usage:        "Audit Database schema",
}

var ArgStripeSecretKey = &cobraviper.StringArgument{
	ArgumentName: "stripe-secret-key",
	EnvName:      "STRIPE_SECRET_KEY",
	Usage:        "Stripe secret key",
}

var ArgPosthogEndpoint = &cobraviper.StringArgument{
	ArgumentName: "posthog-endpoint",
	EnvName:      "POSTHOG_ENDPOINT",
	Usage:        "Posthog endpoint",
}

var ArgPosthogAPIKey = &cobraviper.StringArgument{
	ArgumentName: "posthog-api-key",
	EnvName:      "POSTHOG_API_KEY",
	Usage:        "Posthog API Key",
}

var ArgAnalyticRedisURL = &cobraviper.StringArgument{
	ArgumentName: "analytic-redis-url",
	EnvName:      "ANALYTIC_REDIS_URL",
	Usage:        "Analytic Redis URL",
}

var ArgKubeconfig = &cobraviper.StringArgument{
	ArgumentName: "kubeconfig",
	EnvName:      "KUBECONFIG",
	Usage:        "Path to kubeconfig",
}

var ArgNamespace = &cobraviper.StringArgument{
	ArgumentName: "namespace",
	EnvName:      "NAMESPACE",
	Usage:        "Namespace",
}

var ArgFeatureConfigFilePath = &cobraviper.StringArgument{
	ArgumentName: "file",
	Usage:        "Feature config file path",
}

var ArgPlanName = &cobraviper.StringArgument{
	ArgumentName: "plan-name",
	Usage:        "Plan name",
}

var ArgPlanNameForAppUpdate = &cobraviper.StringArgument{
	ArgumentName: "plan-name",
	Usage:        "Plan name",
	DefaultValue: "custom",
}

var ArgAppHostSuffix = &cobraviper.StringArgument{
	ArgumentName: "app-host-suffix",
	Usage:        "App host suffix",
}

var ArgAnalyticPortalAppID = &cobraviper.StringArgument{
	ArgumentName: "portal-app-id",
	Usage:        "The portal authgear app id",
	DefaultValue: "accounts",
}

var ArgAnalyticPeriod = &cobraviper.StringArgument{
	ArgumentName: "period",
	Usage:        "The period of the report",
}

var ArgAnalyticOutputType = &cobraviper.StringArgument{
	ArgumentName: "output-type",
	Usage:        "Output format of the report, currently supports csv and google-sheets",
	DefaultValue: "csv",
}

var ArgAnalyticCSVOutputFilePath = &cobraviper.StringArgument{
	ArgumentName: "csv-file",
	Usage:        "File path of the output csv file",
}

var ArgAnalyticGoogleOAuthClientCredentialsJSONFilePath = &cobraviper.StringArgument{
	ArgumentName: "google-client-credentials-file",
	Usage:        "File path of client_credentials.json, the file can be downloaded from https://console.developers.google.com, under \"Credentials\"",
	DefaultValue: "./client_credentials.json",
}

var ArgAnalyticGoogleOAuthTokenFilePath = &cobraviper.StringArgument{
	ArgumentName: "google-token-file",
	Usage:        "File path of oauth token file in json format",
	DefaultValue: "./token.json",
}

var ArgAnalyticGoogleSpreadsheetID = &cobraviper.StringArgument{
	ArgumentName: "google-spreadsheet-id",
	Usage:        "The ID of the spreadsheet to update",
}

var ArgAnalyticGoogleSpreadsheetRange = &cobraviper.StringArgument{
	ArgumentName: "google-spreadsheet-range",
	Usage:        "The A1 notation of a range to search for a logical table of data.",
}

var ArgDataJSONFilePath = &cobraviper.StringArgument{
	ArgumentName: "data-json-file",
	Short:        "f",
	Usage:        "File path of the database config source data JSON file",
}

var ArgOutputDirectoryPath = &cobraviper.StringArgument{
	ArgumentName: "output-directory",
	Short:        "o",
	Usage:        "File path of the output directory",
}

var ArgInputDirectoryPath = &cobraviper.StringArgument{
	ArgumentName: "input-directory",
	Short:        "i",
	Usage:        "File path of the input directory",
}

var ArgDefaultDomainSuffix = &cobraviper.StringArgument{
	ArgumentName: "default-domain-suffix",
	Usage:        "e.g. .localhost It must NOT contain a port number.",
}

var ArgDomain = &cobraviper.StringArgument{
	ArgumentName: "domain",
	Usage:        "It must NOT contain a port number.",
}

var ArgApexDomain = &cobraviper.StringArgument{
	ArgumentName: "apex-domain",
	Usage:        "The apex domain of the domain. It must NOT contain a port number.",
}
