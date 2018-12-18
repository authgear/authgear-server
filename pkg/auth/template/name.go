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
)

func VerifyTextTemplateNameForKey(key string) string {
	return fmt.Sprintf("verify_%s_text", key)
}

func VerifyHTMLTemplateNameForKey(key string) string {
	return fmt.Sprintf("verify_%s_html", key)
}
