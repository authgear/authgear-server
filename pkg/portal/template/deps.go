package template

import (
	"github.com/google/wire"

	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/util/template"
)

func NewEngine(defaultDir portalconfig.DefaultTemplateDirectory) *template.Engine {
	// FIXME(template): initialize resolver
	resolver := &template.Resolver{}
	engine := &template.Engine{
		Resolver: resolver,
	}
	return engine
}

var DependencySet = wire.NewSet(
	NewEngine,
)
