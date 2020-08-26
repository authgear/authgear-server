package config

import "github.com/authgear/authgear-server/pkg/util/fs"

type AppContext struct {
	Fs     fs.Fs
	Config *Config
}
