package template

import (
	"path/filepath"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

func RegisterDefaultTemplates(engine *template.Engine, templateDirPath string) {
	templateDir, err := filepath.Abs(templateDirPath)
	if err != nil {
		panic(err)
	}

	engine.RegisterDefaultTemplate(TemplateNameWelcomeEmailText, filepath.Join(templateDir, "welcome_email.txt"))
}

func NewEngineWithConfig(engine *template.Engine, tConfig config.TenantConfiguration) *template.Engine {
	newEngine := template.NewEngineFromEngine(engine)

	if tConfig.WelcomeEmail.TextURL != "" {
		newEngine.RegisterTemplate(TemplateNameWelcomeEmailText, tConfig.WelcomeEmail.TextURL)
	}

	if tConfig.WelcomeEmail.HTMLURL != "" {
		newEngine.RegisterTemplate(TemplateNameWelcomeEmailHTML, tConfig.WelcomeEmail.HTMLURL)
	}

	return newEngine
}
