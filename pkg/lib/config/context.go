package config

import "github.com/authgear/authgear-server/pkg/util/resource"

type AppContext struct {
	AppFs     resource.Fs
	Resources *resource.Manager
	Config    *Config
}
