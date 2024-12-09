package config

import "github.com/authgear/authgear-server/pkg/util/resource"

type AppDomains []string

type AppContext struct {
	AppFs     resource.Fs
	PlanFs    resource.Fs
	Resources *resource.Manager
	Config    *Config
	PlanName  string
	Domains   AppDomains
}
