package template

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

func RegisterDefaultTemplates(engine *template.Engine) {
	// TODO(template)
}

// NewEngineWithConfig return new engine with loaders from the config
// nolint: gocyclo
func NewEngineWithConfig(engine *template.Engine, tConfig config.TenantConfiguration) *template.Engine {
	newEngine := template.NewEngine()
	engine.CopyDefaultToEngine(newEngine)
	loader := template.NewHTTPLoader()
	// TODO(template)
	newEngine.SetLoaders([]template.Loader{loader})
	return newEngine
}
