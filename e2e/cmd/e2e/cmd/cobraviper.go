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

var ArgElasticsearchURL = &cobraviper.StringArgument{
	ArgumentName: "elasticsearch-url",
	EnvName:      "ELASTICSEARCH_URL",
	Usage:        "Elasticsearch URL",
}

var ArgAppID = &cobraviper.StringArgument{
	ArgumentName: "app-id",
	EnvName:      "APP_ID",
	Usage:        "App ID",
}

var ArgOutputFolder = &cobraviper.StringArgument{
	ArgumentName: "output-folder",
	EnvName:      "OUTPUT_FOLDER",
	Short:        "o",
	Usage:        "Path to output folder",
}

var ArgInputFolder = &cobraviper.StringArgument{
	ArgumentName: "input-folder",
	EnvName:      "INPUT_FOLDER",
	Short:        "i",
	Usage:        "Path to input folder",
}

var ArgConfigSource = &cobraviper.StringArgument{
	ArgumentName: "config-source",
	Usage:        "Config source",
}

var ArgConfigOverride = &cobraviper.StringArgument{
	ArgumentName: "config-override",
	Usage:        "Config override",
}

var ArgConfigSourceExtraFilesDirectory = &cobraviper.StringArgument{
	ArgumentName: "config-source-extra-files-directory",
	Usage:        "Config source extra files directory",
}

var ArgCustomSQL = &cobraviper.StringArgument{
	ArgumentName: "custom-sql",
	Usage:        "Filepath to custom sql",
}

var ArgRawSQL = &cobraviper.StringArgument{
	ArgumentName: "raw-sql",
	Usage:        "Custom raw SQL statements",
}

var ArgSelectUserIDSQL = &cobraviper.StringArgument{
	ArgumentName: "select-user-id-sql",
	Usage:        "A SQL to select a user id. The result of the SQL must be one single row with one single column, which is the user ID.",
}

var ArgSessionType = &cobraviper.StringArgument{
	ArgumentName: "session-type",
	Usage:        "The session type. idp or offline_grant.",
}

var ArgSessionID = &cobraviper.StringArgument{
	ArgumentName: "session-id",
	Usage:        "The session id.",
}

var ArgToken = &cobraviper.StringArgument{
	ArgumentName: "token",
	Usage:        "A token.",
}

var ArgClientID = &cobraviper.StringArgument{
	ArgumentName: "client-id",
	Usage:        "Client ID.",
}

var ArgPurpose = &cobraviper.StringArgument{
	ArgumentName: "purpose",
	Usage:        "The purpose of challenge.",
}
