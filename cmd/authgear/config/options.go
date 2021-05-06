package config

import "github.com/authgear/authgear-server/pkg/lib/config"

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

	opts.ElasticsearchURL = promptString{
		Title:        "Elasticsearch URL",
		DefaultValue: "http://localhost:9200",
	}.Prompt()

	opts.RedisURL = promptURL{
		Title:        "Redis URL",
		DefaultValue: "redis://localhost",
	}.Prompt()

	return opts
}
