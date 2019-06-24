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

// NewEngineWithConfig return new engine with loaders from the config
// nolint: gocyclo
func NewEngineWithConfig(engine *template.Engine, tConfig config.TenantConfiguration) *template.Engine {
	newEngine := template.NewEngine()
	engine.CopyDefaultToEngine(newEngine)

	loader := template.NewHTTPLoader()

	if tConfig.UserConfig.WelcomeEmail.TextURL != "" {
		loader.URLMap[TemplateNameWelcomeEmailText] = tConfig.UserConfig.WelcomeEmail.TextURL
	}

	if tConfig.UserConfig.WelcomeEmail.HTMLURL != "" {
		loader.URLMap[TemplateNameWelcomeEmailHTML] = tConfig.UserConfig.WelcomeEmail.HTMLURL
	}

	if tConfig.UserConfig.ForgotPassword.EmailTextURL != "" {
		loader.URLMap[TemplateNameForgotPasswordEmailText] = tConfig.UserConfig.ForgotPassword.EmailTextURL
	}

	if tConfig.UserConfig.ForgotPassword.EmailHTMLURL != "" {
		loader.URLMap[TemplateNameForgotPasswordEmailHTML] = tConfig.UserConfig.ForgotPassword.EmailHTMLURL
	}

	if tConfig.UserConfig.UserVerification.ErrorHTMLURL != "" {
		loader.URLMap[TemplateNameVerifyErrorHTML] = tConfig.UserConfig.UserVerification.ErrorHTMLURL
	}

	for _, keyConfig := range tConfig.UserConfig.UserVerification.Keys {
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

	if tConfig.UserConfig.ForgotPassword.ResetErrorHTMLURL != "" {
		loader.URLMap[TemplateNameResetPasswordErrorHTML] = tConfig.UserConfig.ForgotPassword.ResetErrorHTMLURL
	}

	if tConfig.UserConfig.ForgotPassword.ResetSuccessHTMLURL != "" {
		loader.URLMap[TemplateNameResetPasswordSuccessHTML] = tConfig.UserConfig.ForgotPassword.ResetSuccessHTMLURL
	}

	if tConfig.UserConfig.ForgotPassword.ResetHTMLURL != "" {
		loader.URLMap[TemplateNameResetPasswordHTML] = tConfig.UserConfig.ForgotPassword.ResetHTMLURL
	}

	newEngine.SetLoaders([]template.Loader{loader})
	return newEngine
}
