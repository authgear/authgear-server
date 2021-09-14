package main

import (
	"github.com/authgear/authgear-server/pkg/util/cobraviper"
)

func newInt(v int) *int { return &v }

var cvbinder *cobraviper.Binder

func getBinder() *cobraviper.Binder {
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

var ArgDefaultAuthgearDomain = &cobraviper.StringArgument{
	ArgumentName: "default-authgear-domain",
	EnvName:      "DEFAULT_AUTHGEAR_DOMAIN",
	Usage:        "App default domain",
}

var ArgCustomAuthgearDomain = &cobraviper.StringArgument{
	ArgumentName: "custom-authgear-domain",
	EnvName:      "CUSTOM_AUTHGEAR_DOMAIN",
	Usage:        "App custom domain",
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

var ArgAnalyticYear = &cobraviper.IntArgument{
	ArgumentName: "year",
	Usage:        "Year of the report",
	EnvName:      "YEAR",
	Min:          newInt(1900),
	Max:          newInt(3000),
}

var ArgAnalyticISOWeek = &cobraviper.IntArgument{
	ArgumentName: "week",
	Usage:        "ISO week of the weekly report",
	Min:          newInt(1),
	Max:          newInt(53),
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
