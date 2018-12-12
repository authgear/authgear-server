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
}

func NewEngineWithConfig(engine *template.Engine, tConfig config.TenantConfiguration) *template.Engine {
	newEngine := template.NewEngine()
	engine.CopyDefaultToEngine(newEngine)

	loader := template.NewHTTPLoader()

	if tConfig.WelcomeEmail.TextURL != "" {
		loader.URLMap[TemplateNameWelcomeEmailText] = tConfig.WelcomeEmail.TextURL
	}

	if tConfig.WelcomeEmail.HTMLURL != "" {
		loader.URLMap[TemplateNameWelcomeEmailHTML] = tConfig.WelcomeEmail.HTMLURL
	}

	if tConfig.ForgotPassword.EmailTextURL != "" {
		loader.URLMap[TemplateNameForgotPasswordEmailText] = tConfig.ForgotPassword.EmailTextURL
	}

	if tConfig.ForgotPassword.EmailHTMLURL != "" {
		loader.URLMap[TemplateNameForgotPasswordEmailHTML] = tConfig.ForgotPassword.EmailHTMLURL
	}

	for _, keyConfig := range tConfig.UserVerify.KeyConfigs {
		providerConfig := keyConfig.ProviderConfig
		if providerConfig.TextURL != "" {
			loader.URLMap[VerifyTextTemplateNameForKey(keyConfig.Key)] = providerConfig.TextURL
		}

		if providerConfig.HTMLURL != "" {
			loader.URLMap[VerifyHTMLTemplateNameForKey(keyConfig.Key)] = providerConfig.HTMLURL
		}
	}

	return newEngine
}
