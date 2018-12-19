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

	if tConfig.UserVerify.ErrorHTMLURL != "" {
		loader.URLMap[TemplateNameVerifyErrorHTML] = tConfig.UserVerify.ErrorHTMLURL
	}

	for _, keyConfig := range tConfig.UserVerify.KeyConfigs {
		if keyConfig.SuccessHTMLURL != "" {
			loader.URLMap[VerifySuccessHTMLTemplateNameForKey(keyConfig.Key)] = keyConfig.SuccessHTMLURL
		}

		if keyConfig.ErrorHTMLURL != "" {
			loader.URLMap[VerifyErrorHTMLTemplateNameForKey(keyConfig.Key)] = keyConfig.ErrorHTMLURL
		}

		providerConfig := keyConfig.ProviderConfig
		if providerConfig.TextURL != "" {
			loader.URLMap[VerifyTextTemplateNameForKey(keyConfig.Key)] = providerConfig.TextURL
		}

		if providerConfig.HTMLURL != "" {
			loader.URLMap[VerifyHTMLTemplateNameForKey(keyConfig.Key)] = providerConfig.HTMLURL
		}
	}

	if tConfig.ForgotPassword.ResetErrorHTMLURL != "" {
		loader.URLMap[TemplateNameResetPasswordErrorHTML] = tConfig.ForgotPassword.ResetErrorHTMLURL
	}

	if tConfig.ForgotPassword.ResetSuccessHTMLURL != "" {
		loader.URLMap[TemplateNameResetPasswordSuccessHTML] = tConfig.ForgotPassword.ResetSuccessHTMLURL
	}

	if tConfig.ForgotPassword.ResetHTMLURL != "" {
		loader.URLMap[TemplateNameResetPasswordHTML] = tConfig.ForgotPassword.ResetHTMLURL
	}

	newEngine.SetLoaders([]template.Loader{loader})
	return newEngine
}
