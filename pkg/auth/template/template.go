package template

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/forgotpassword"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/welcemail"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

const DefaultErrorHTML = `<!DOCTYPE html>
<head>
  <meta name="viewport" content="width=device-width, initial-scale=1">
</head>
<p>{{ .error.Message }}</p>`

func NewEngineWithConfig(
	tConfig config.TenantConfiguration,
	enableFileSystemTemplate bool,
	assetGearLoader *template.AssetGearLoader,
) *template.Engine {
	e := template.NewEngine(template.NewEngineOptions{
		EnableFileLoader: enableFileSystemTemplate,
		TemplateItems:    tConfig.TemplateItems,
		AssetGearLoader:  assetGearLoader,
		FallbackLanguage: tConfig.AppConfig.Localization.FallbackLanguage,
	})

	e.Register(welcemail.TemplateWelcomeEmailTXT)
	e.Register(welcemail.TemplateWelcomeEmailHTML)

	e.Register(userverify.TemplateUserVerificationSMSTXT)
	e.Register(userverify.TemplateUserVerificationEmailTXT)
	e.Register(userverify.TemplateUserVerificationEmailHTML)
	e.Register(userverify.TemplateUserVerificationSuccessHTML)
	e.Register(userverify.TemplateUserVerificationErrorHTML)

	e.Register(mfa.TemplateMFAOOBCodeSMSTXT)
	e.Register(mfa.TemplateMFAOOBCodeEmailTXT)
	e.Register(mfa.TemplateMFAOOBCodeEmailHTML)

	e.Register(interaction.TemplateOOBCodeSMSTXT)
	e.Register(interaction.TemplateOOBCodeEmailTXT)
	e.Register(interaction.TemplateOOBCodeEmailHTML)

	// Auth UI
	e.Register(webapp.TemplateAuthUITranslationJSON)

	e.Register(webapp.TemplateAuthUIHTMLHeadHTML)
	e.Register(webapp.TemplateAuthUIHeaderHTML)
	e.Register(webapp.TemplateAuthUIFooterHTML)

	e.Register(webapp.TemplateAuthUILoginHTML)
	e.Register(webapp.TemplateAuthUILoginPasswordHTML)
	e.Register(webapp.TemplateAuthUIForgotPasswordHTML)
	e.Register(webapp.TemplateAuthUIForgotPasswordSuccessHTML)
	e.Register(webapp.TemplateAuthUIResetPasswordHTML)
	e.Register(webapp.TemplateAuthUIResetPasswordSuccessHTML)
	e.Register(webapp.TemplateAuthUISignupHTML)
	e.Register(webapp.TemplateAuthUISignupPasswordHTML)
	e.Register(webapp.TemplateAuthUISettingsHTML)
	e.Register(webapp.TemplateAuthUILogoutHTML)

	e.Register(forgotpassword.TemplateForgotPasswordEmailTXT)
	e.Register(forgotpassword.TemplateForgotPasswordEmailHTML)
	e.Register(forgotpassword.TemplateForgotPasswordSMSTXT)

	return e
}
