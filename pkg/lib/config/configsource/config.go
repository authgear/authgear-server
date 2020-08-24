package configsource

type Type string

const (
	TypeLocalFS Type = "local_fs"
)

var Types = []Type{
	TypeLocalFS,
}

type Config struct {
	// Type sets the type of configuration source
	Type Type `envconfig:"TYPE" default:"local_fs"`

	// Watch indicates whether the configuration source would watch for changes and reload automatically
	Watch bool `envconfig:"WATCH" default:"true"`
	// Directory sets the path to app configuration directory file for local FS sources
	Directory string `envconfig:"DIRECTORY" default:"."`
}
