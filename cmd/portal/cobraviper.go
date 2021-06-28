package main

import (
	"github.com/authgear/authgear-server/pkg/util/cobraviper"
)

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
