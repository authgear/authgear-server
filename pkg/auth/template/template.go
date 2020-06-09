package template

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/oob"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/forgotpassword"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/welcomemessage"
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

	e.Register(welcomemessage.TemplateWelcomeEmailTXT)
	e.Register(welcomemessage.TemplateWelcomeEmailHTML)

	e.Register(oob.TemplateOOBCodeSMSTXT)
	e.Register(oob.TemplateOOBCodeEmailTXT)
	e.Register(oob.TemplateOOBCodeEmailHTML)

	// Auth UI
	e.Register(webapp.TemplateAuthUITranslationJSON)

	e.Register(webapp.TemplateAuthUIHTMLHeadHTML)
	e.Register(webapp.TemplateAuthUIHeaderHTML)
	e.Register(webapp.TemplateAuthUIFooterHTML)

	e.Register(webapp.TemplateAuthUILoginHTML)
	e.Register(webapp.TemplateAuthUISignupHTML)
	e.Register(webapp.TemplateAuthUIPromoteHTML)

	e.Register(webapp.TemplateAuthUIEnterPasswordHTML)
	e.Register(webapp.TemplateAuthUICreatePasswordHTML)
	e.Register(webapp.TemplateAuthUIOOBOTPHTML)
	e.Register(webapp.TemplateAuthUIEnterLoginIDHTML)

	e.Register(webapp.TemplateAuthUIForgotPasswordHTML)
	e.Register(webapp.TemplateAuthUIForgotPasswordSuccessHTML)
	e.Register(webapp.TemplateAuthUIResetPasswordHTML)
	e.Register(webapp.TemplateAuthUIResetPasswordSuccessHTML)
	e.Register(webapp.TemplateAuthUILogoutHTML)

	e.Register(webapp.TemplateAuthUISettingsHTML)
	e.Register(webapp.TemplateAuthUISettingsIdentityHTML)

	e.Register(forgotpassword.TemplateForgotPasswordEmailTXT)
	e.Register(forgotpassword.TemplateForgotPasswordEmailHTML)
	e.Register(forgotpassword.TemplateForgotPasswordSMSTXT)

	return e
}
