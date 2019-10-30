package template

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

func RegisterDefaultTemplates(engine *template.Engine) {
	engine.RegisterDefaultTemplate(TemplateNameWelcomeEmailText, templateWelcomeEmailTxt)
	engine.RegisterDefaultTemplate(TemplateNameForgotPasswordEmailText, templateForgotPasswordEmailTxt)
	engine.RegisterDefaultTemplate(TemplateNameVerifyEmailText, templateVerifyEmailTxt)
	engine.RegisterDefaultTemplate(TemplateNameVerifySMSText, templateVerifySMSTxt)
	engine.RegisterDefaultTemplate(TemplateNameResetPasswordErrorHTML, templateResetPasswordErrorHTML)
	engine.RegisterDefaultTemplate(TemplateNameResetPasswordSuccessHTML, templateResetPasswordSuccessHTML)
	engine.RegisterDefaultTemplate(TemplateNameResetPasswordHTML, templateResetPasswordHTML)
	engine.RegisterDefaultTemplate(TemplateNameVerifyErrorHTML, templateVerifyErrorHTML)
	engine.RegisterDefaultTemplate(TemplateNameVerifySuccessHTML, templateVerifySuccessHTML)
	// MFA
	engine.RegisterDefaultTemplate(TemplateNameMFAOOBCodeSMSText, templateMFAOOBCodeSMSText)
	engine.RegisterDefaultTemplate(TemplateNameMFAOOBCodeEmailText, templateMFAOOBCodeEmailText)
	engine.RegisterDefaultTemplate(TemplateNameMFAOOBCodeEmailHTML, templateMFAOOBCodeEmailHTML)
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
