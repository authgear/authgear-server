package template

import (
	"fmt"
)

const (
	// TemplateNameWelcomeEmailText is the template name of welcome email text
	TemplateNameWelcomeEmailText = "welcome_email_text"

	// TemplateNameWelcomeEmailHTML is the template name of welcome email html
	TemplateNameWelcomeEmailHTML = "welcome_email_html"

	// TemplateNameForgotPasswordEmailText is the template name of forgot password email text
	TemplateNameForgotPasswordEmailText = "forgot_password_email_text"

	// TemplateNameForgotPasswordEmailHTML is the template name of forgot password email html
	TemplateNameForgotPasswordEmailHTML = "forgot_password_email_html"

	// TemplateNameVerifyEmailText is the template name of verify email text
	TemplateNameVerifyEmailText = "verify_email_text"

	// TemplateNameVerifyEmailHTML is the template name of verify email html
	TemplateNameVerifyEmailHTML = "verify_email_html"

	// TemplateNameVerifySMSText is the template name of verify sms text
	TemplateNameVerifySMSText = "verify_sms_text"

	// TemplateNameResetPasswordErrorHTML is the template name of reset password error html
	TemplateNameResetPasswordErrorHTML = "reset_password_error_html"

	// TemplateNameResetPasswordSuccessHTML is the template name of reset password success html
	TemplateNameResetPasswordSuccessHTML = "reset_password_success_html"

	// TemplateNameResetPasswordHTML is the template name of reset password html
	TemplateNameResetPasswordHTML = "reset_password_html"
)

func VerifyTextTemplateNameForKey(key string) string {
	return fmt.Sprintf("verify_%s_text", key)
}

func VerifyHTMLTemplateNameForKey(key string) string {
	return fmt.Sprintf("verify_%s_html", key)
}
