package config

type Options struct {
	AppID        string
	PublicOrigin string
}

func ReadOptionsFromConsole() *Options {
	opts := &Options{}

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

type SecretOptions struct {
	DatabaseURL    string
	DatabaseSchema string
	RedisURL       string
}

func ReadSecretOptionsFromConsole() *SecretOptions {
	opts := &SecretOptions{}

	opts.DatabaseURL = promptURL{
		Title:        "Database URL",
		DefaultValue: "postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable",
	}.Prompt()

	opts.DatabaseSchema = promptString{
		Title:        "Database schema",
		DefaultValue: "public",
	}.Prompt()

	opts.RedisURL = promptURL{
		Title:        "Redis URL",
		DefaultValue: "redis://localhost",
	}.Prompt()

	return opts
}
