package deps

import (
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
	"github.com/authgear/authgear-server/pkg/lib/feature/welcomemessage"
	"github.com/authgear/authgear-server/pkg/util/template"
)

func NewEngineWithConfig(
	serverConfig *config.ServerConfig,
	c *config.Config,
) *template.Engine {
	var refs []template.Reference
	for _, item := range c.AppConfig.Template.Items {
		refs = append(refs, template.Reference{
			Type:        string(item.Type),
			LanguageTag: item.LanguageTag,
			URI:         item.URI,
		})
	}

	resolver := template.NewResolver(template.NewResolverOptions{
		DefaultTemplatesDirectory: serverConfig.DefaultTemplateDirectory,
		References:                refs,
		FallbackLanguageTag:       c.AppConfig.Localization.FallbackLanguage,
	})
	engine := &template.Engine{
		Resolver: resolver,
	}

	resolver.Register(welcomemessage.TemplateWelcomeEmailTXT)
	resolver.Register(welcomemessage.TemplateWelcomeEmailHTML)

	resolver.Register(otp.TemplateVerificationSMSTXT)
	resolver.Register(otp.TemplateVerificationEmailTXT)
	resolver.Register(otp.TemplateVerificationEmailHTML)
	resolver.Register(otp.TemplateSetupPrimaryOOBSMSTXT)
	resolver.Register(otp.TemplateSetupPrimaryOOBEmailTXT)
	resolver.Register(otp.TemplateSetupPrimaryOOBEmailHTML)
	resolver.Register(otp.TemplateSetupSecondaryOOBSMSTXT)
	resolver.Register(otp.TemplateSetupSecondaryOOBEmailTXT)
	resolver.Register(otp.TemplateSetupSecondaryOOBEmailHTML)
	resolver.Register(otp.TemplateAuthenticatePrimaryOOBSMSTXT)
	resolver.Register(otp.TemplateAuthenticatePrimaryOOBEmailTXT)
	resolver.Register(otp.TemplateAuthenticatePrimaryOOBEmailHTML)
	resolver.Register(otp.TemplateAuthenticateSecondaryOOBSMSTXT)
	resolver.Register(otp.TemplateAuthenticateSecondaryOOBEmailTXT)
	resolver.Register(otp.TemplateAuthenticateSecondaryOOBEmailHTML)

	// Auth UI
	resolver.Register(webapp.TemplateAuthUITranslationJSON)

	resolver.Register(webapp.TemplateAuthUIHTMLHeadHTML)
	resolver.Register(webapp.TemplateAuthUIHeaderHTML)
	resolver.Register(webapp.TemplateAuthUIFooterHTML)

	resolver.Register(webapp.TemplateAuthUILoginHTML)
	resolver.Register(webapp.TemplateAuthUISignupHTML)
	resolver.Register(webapp.TemplateAuthUIPromoteHTML)

	resolver.Register(webapp.TemplateAuthUIEnterPasswordHTML)
	resolver.Register(webapp.TemplateAuthUICreatePasswordHTML)
	resolver.Register(webapp.TemplateAuthUISetupTOTPHTML)
	resolver.Register(webapp.TemplateAuthUIEnterTOTPHTML)
	resolver.Register(webapp.TemplateAuthUISetupOOBOTPHTML)
	resolver.Register(webapp.TemplateAuthUIEnterOOBOTPHTML)
	resolver.Register(webapp.TemplateAuthUIEnterLoginIDHTML)
	resolver.Register(webapp.TemplateAuthUIEnterRecoveryCodeHTML)
	resolver.Register(webapp.TemplateAuthUISetupRecoveryCodeHTML)
	resolver.Register(webapp.TemplateAuthUIDownloadRecoveryCodeTXT)
	resolver.Register(webapp.TemplateAuthUIVerifyIdentityHTML)
	resolver.Register(webapp.TemplateAuthUIVerifyIdentitySuccessHTML)

	resolver.Register(webapp.TemplateAuthUIForgotPasswordHTML)
	resolver.Register(webapp.TemplateAuthUIForgotPasswordSuccessHTML)
	resolver.Register(webapp.TemplateAuthUIResetPasswordHTML)
	resolver.Register(webapp.TemplateAuthUIResetPasswordSuccessHTML)
	resolver.Register(webapp.TemplateAuthUILogoutHTML)

	resolver.Register(webapp.TemplateAuthUISettingsHTML)
	resolver.Register(webapp.TemplateAuthUISettingsIdentityHTML)

	resolver.Register(forgotpassword.TemplateForgotPasswordEmailTXT)
	resolver.Register(forgotpassword.TemplateForgotPasswordEmailHTML)
	resolver.Register(forgotpassword.TemplateForgotPasswordSMSTXT)

	return engine
}
