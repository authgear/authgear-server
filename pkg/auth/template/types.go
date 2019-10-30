package template

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
)

const (
	TemplateItemTypeForgotPasswordEmailTXT           config.TemplateItemType = "forgot_password_email.txt"
	TemplateItemTypeForgotPasswordEmailHTML          config.TemplateItemType = "forgot_password_email.html"
	TemplateItemTypeForgotPasswordResetHTML          config.TemplateItemType = "forgot_password_reset.html"
	TemplateItemTypeForgotPasswordSuccessHTML        config.TemplateItemType = "forgot_password_success.html"
	TemplateItemTypeForgotPasswordErrorHTML          config.TemplateItemType = "forgot_password_error.html"
	TemplateItemTypeWelcomeEmailTXT                  config.TemplateItemType = "welcome_email.txt"
	TemplateItemTypeWelcomeEmailHTML                 config.TemplateItemType = "welcome_email.html"
	TemplateItemTypeUserVerificationGeneralErrorHTML config.TemplateItemType = "user_verification_general_error.html"
	TemplateItemTypeUserVerificationSMSTXT           config.TemplateItemType = "user_verification_sms.txt"
	TemplateItemTypeUserVerificationEmailTXT         config.TemplateItemType = "user_verification_email.txt"
	TemplateItemTypeUserVerificationEmailHTML        config.TemplateItemType = "user_verification_email.html"
	TemplateItemTypeUserVerificationSuccessHTML      config.TemplateItemType = "user_verification_success.html"
	TemplateItemTypeUserVerificationErrorHTML        config.TemplateItemType = "user_verification_error.html"
	TemplateItemTypeMFAOOBCodeSMSTXT                 config.TemplateItemType = "mfa_oob_code_sms.txt"
	TemplateItemTypeMFAOOBCodeEmailTXT               config.TemplateItemType = "mfa_oob_code_email.txt"
	TemplateItemTypeMFAOOBCodeEmailHTML              config.TemplateItemType = "mfa_oob_code_email.html"
)
