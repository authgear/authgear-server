package template

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

func RegisterDefaultTemplates(engine *template.Engine) {
	engine.RegisterDefaultTemplate(TemplateNameWelcomeEmailText, templateWelcomeEmailTxt)
	engine.RegisterDefaultTemplate(TemplateNameForgotPasswordEmailText, templateForgotPasswordEmailTxt)
}

func NewEngineWithConfig(engine *template.Engine, tConfig config.TenantConfiguration) *template.Engine {
	newEngine := template.NewEngine()
	engine.CopyDefaultToEngine(newEngine)

	loader := template.NewHTTPLoader()

	if tConfig.WelcomeEmail.TextURL != "" {
		loader.UrlMap[TemplateNameWelcomeEmailText] = tConfig.WelcomeEmail.TextURL
	}

	if tConfig.WelcomeEmail.HTMLURL != "" {
		loader.UrlMap[TemplateNameWelcomeEmailHTML] = tConfig.WelcomeEmail.HTMLURL
	}

	if tConfig.ForgotPassword.EmailTextURL != "" {
		loader.UrlMap[TemplateNameForgotPasswordEmailText] = tConfig.ForgotPassword.EmailTextURL
	}

	if tConfig.ForgotPassword.EmailHTMLURL != "" {
		loader.UrlMap[TemplateNameForgotPasswordEmailHTML] = tConfig.ForgotPassword.EmailHTMLURL
	}

	return newEngine
}
