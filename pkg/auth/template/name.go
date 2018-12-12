package template

import (
	"fmt"
)

const (
	TemplateNameWelcomeEmailText        = "welcome_email_text"
	TemplateNameWelcomeEmailHTML        = "welcome_email_html"
	TemplateNameForgotPasswordEmailText = "forgot_password_email_text"
	TemplateNameForgotPasswordEmailHTML = "forgot_password_email_html"
	TemplateNameVerifyEmailText         = "verify_email_text"
	TemplateNameVerifyEmailHTML         = "verify_email_html"
	TemplateNameVerifySMSText           = "verify_sms_text"
)

func VerifyTextTemplateNameForKey(key string) string {
	return fmt.Sprintf("verify_%s_text", key)
}

func VerifyHTMLTemplateNameForKey(key string) string {
	return fmt.Sprintf("verify_%s_html", key)
}
