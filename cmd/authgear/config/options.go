package config

import "strings"

type Options struct {
	AppID string
	Hosts []string
}

func ReadOptionsFromConsole() *Options {
	opts := &Options{}

	opts.AppID = promptString{
		Title:        "App ID",
		DefaultValue: "my-app",
	}.Prompt()

	opts.Hosts = strings.Split(promptString{
		Title:        "Host name of authgear",
		DefaultValue: "localhost:3000",
	}.Prompt(), ";")

	return opts
}

type SecretOptions struct {
	DatabaseURL    string
	DatabaseSchema string
	RedisHost      string
	RedisPort      int
	RedisPassword  string
}

func ReadSecretOptionsFromConsole() *SecretOptions {
	opts := &SecretOptions{}

	opts.DatabaseURL = promptURL{
		Title:        "Database URL",
		DefaultValue: "postgres://postgres@127.0.0.1:5432/postgres?sslmode=disable",
	}.Prompt()

	opts.DatabaseSchema = promptString{
		Title:        "Database schema",
		DefaultValue: "public",
	}.Prompt()

	opts.RedisHost = promptURL{
		Title:        "Redis host",
		DefaultValue: "localhost",
	}.Prompt()

	opts.RedisPort = promptInt{
		Title:        "Redis port",
		DefaultValue: 6379,
	}.Prompt()

	opts.RedisPassword = promptString{
		Title:        "Redis password",
		DefaultValue: "",
	}.Prompt()

	return opts
}
