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

	// TemplateNameVerifySuccessHTML is the template name of verify success html
	TemplateNameVerifySuccessHTML = "verify_success_html"

	// TemplateNameVerifyErrorHTML is the template name of verify error html
	TemplateNameVerifyErrorHTML = "verify_error_html"
)

func VerifyTemplateNameForKey(key string, templateType string) string {
	return fmt.Sprintf("verify_%s_%s", key, templateType)
}

func VerifyTextTemplateNameForKey(key string) string {
	return VerifyTemplateNameForKey(key, "text")
}

func VerifyHTMLTemplateNameForKey(key string) string {
	return VerifyTemplateNameForKey(key, "html")
}

func VerifySuccessHTMLTemplateNameForKey(key string) string {
	return fmt.Sprintf("%s_verify_success_html", key)
}

func VerifyErrorHTMLTemplateNameForKey(key string) string {
	return fmt.Sprintf("%s_verify_error_html", key)
}
