package template

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/forgotpassword"
	"github.com/authgear/authgear-server/pkg/auth/dependency/welcomemessage"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/otp"
	"github.com/authgear/authgear-server/pkg/template"
)

func NewEngineWithConfig(
	serverConfig *config.ServerConfig,
	c *config.Config,
) *template.Engine {
	e := template.NewEngine(template.NewEngineOptions{
		DefaultTemplatesDirectory: serverConfig.DefaultTemplateDirectory,
		TemplateItems:             c.AppConfig.Template.Items,
		FallbackLanguage:          c.AppConfig.Localization.FallbackLanguage,
	})

	e.Register(welcomemessage.TemplateWelcomeEmailTXT)
	e.Register(welcomemessage.TemplateWelcomeEmailHTML)

	e.Register(otp.TemplateVerificationSMSTXT)
	e.Register(otp.TemplateVerificationEmailTXT)
	e.Register(otp.TemplateVerificationEmailHTML)
	e.Register(otp.TemplateSetupPrimaryOOBSMSTXT)
	e.Register(otp.TemplateSetupPrimaryOOBEmailTXT)
	e.Register(otp.TemplateSetupPrimaryOOBEmailHTML)
	e.Register(otp.TemplateSetupSecondaryOOBSMSTXT)
	e.Register(otp.TemplateSetupSecondaryOOBEmailTXT)
	e.Register(otp.TemplateSetupSecondaryOOBEmailHTML)
	e.Register(otp.TemplateAuthenticatePrimaryOOBSMSTXT)
	e.Register(otp.TemplateAuthenticatePrimaryOOBEmailTXT)
	e.Register(otp.TemplateAuthenticatePrimaryOOBEmailHTML)
	e.Register(otp.TemplateAuthenticateSecondaryOOBSMSTXT)
	e.Register(otp.TemplateAuthenticateSecondaryOOBEmailTXT)
	e.Register(otp.TemplateAuthenticateSecondaryOOBEmailHTML)

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
	e.Register(webapp.TemplateAuthUISetupTOTPHTML)
	e.Register(webapp.TemplateAuthUIEnterTOTPHTML)
	e.Register(webapp.TemplateAuthUISetupOOBOTPHTML)
	e.Register(webapp.TemplateAuthUIEnterOOBOTPHTML)
	e.Register(webapp.TemplateAuthUIEnterLoginIDHTML)
	e.Register(webapp.TemplateAuthUISetupRecoveryCodeHTML)
	e.Register(webapp.TemplateAuthUIDownloadRecoveryCodeTXT)
	e.Register(webapp.TemplateAuthUIVerifyIdentityHTML)
	e.Register(webapp.TemplateAuthUIVerifyIdentitySuccessHTML)

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
