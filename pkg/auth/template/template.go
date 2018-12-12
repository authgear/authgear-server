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

	for _, keyConfig := range tConfig.UserVerify.KeyConfigs {
		providerConfig := keyConfig.ProviderConfig
		if providerConfig.TextURL != "" {
			loader.UrlMap[VerifyTextTemplateNameForKey(keyConfig.Key)] = providerConfig.TextURL
		}

		if providerConfig.HTMLURL != "" {
			loader.UrlMap[VerifyHTMLTemplateNameForKey(keyConfig.Key)] = providerConfig.HTMLURL
		}
	}

	return newEngine
}
