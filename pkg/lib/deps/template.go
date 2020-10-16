package deps

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/fs"
	"github.com/authgear/authgear-server/pkg/util/template"
)

func NewEngineWithConfig(
	appFs fs.Fs,
	defaultDir string,
	c *config.Config,
) *template.Engine {
	// FIXME(template): initialize resolver
	resolver := &template.Resolver{}
	engine := &template.Engine{
		Resolver: resolver,
	}

	return engine
}
