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
	Usage:        "Path to output folder",
}
