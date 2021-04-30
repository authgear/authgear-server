package main

import (
	"github.com/authgear/authgear-server/pkg/util/cobraviper"
)

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

var ArgOutput = &cobraviper.StringArgument{
	ArgumentName: "output",
	EnvName:      "OUTPUT",
	Short:        "o",
	Usage:        "Path to output",
}

var ArgElasticsearchURL = &cobraviper.StringArgument{
	ArgumentName: "elasticsearch-url",
	EnvName:      "ELASTICSEARCH_URL",
	Usage:        "Elasticsearch URL",
}
