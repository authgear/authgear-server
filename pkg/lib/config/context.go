package config

import "github.com/authgear/authgear-server/pkg/util/resource"

type AppContext struct {
	Resources *resource.Manager
	Config    *Config
}
