package otp

import (
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/util/template"
)

type MessageType string

const (
	MessageTypeVerification             MessageType = "verification"
	MessageTypeSetupPrimaryOOB          MessageType = "setup-primary-oob"
	MessageTypeSetupSecondaryOOB        MessageType = "setup-secondary-oob"
	MessageTypeAuthenticatePrimaryOOB   MessageType = "authenticate-primary-oob"
	MessageTypeAuthenticateSecondaryOOB MessageType = "authenticate-secondary-oob"
	MessageTypeForgotPassword           MessageType = "forgot-password"
	MessageTypeWhatsappCode             MessageType = "whatsapp-code"
)

type messageTemplateContext struct {
	Email string
	Phone string
	Code  string
	URL   string
	Host  string

	// compatibility with forgot password templates
	Link string
}

var (
	TemplateMessageVerificationSMSTXT    = template.RegisterMessagePlainText("messages/verification_sms.txt")
	TemplateMessageVerificationEmailTXT  = template.RegisterMessagePlainText("messages/verification_email.txt")
	TemplateMessageVerificationEmailHTML = template.RegisterMessageHTML("messages/verification_email.html")

	TemplateMessageSetupPrimaryOOBSMSTXT    = template.RegisterMessagePlainText("messages/setup_primary_oob_sms.txt")
	TemplateMessageSetupPrimaryOOBEmailTXT  = template.RegisterMessagePlainText("messages/setup_primary_oob_email.txt")
	TemplateMessageSetupPrimaryOOBEmailHTML = template.RegisterMessageHTML("messages/setup_primary_oob_email.html")

	TemplateMessageSetupSecondaryOOBSMSTXT    = template.RegisterMessagePlainText("messages/setup_secondary_oob_sms.txt")
	TemplateMessageSetupSecondaryOOBEmailTXT  = template.RegisterMessagePlainText("messages/setup_secondary_oob_email.txt")
	TemplateMessageSetupSecondaryOOBEmailHTML = template.RegisterMessageHTML("messages/setup_secondary_oob_email.html")

	TemplateMessageAuthenticatePrimaryOOBSMSTXT    = template.RegisterMessagePlainText("messages/authenticate_primary_oob_sms.txt")
	TemplateMessageAuthenticatePrimaryOOBEmailTXT  = template.RegisterMessagePlainText("messages/authenticate_primary_oob_email.txt")
	TemplateMessageAuthenticatePrimaryOOBEmailHTML = template.RegisterMessageHTML("messages/authenticate_primary_oob_email.html")

	TemplateMessageAuthenticatePrimaryLoginLinkEmailTXT  = template.RegisterMessagePlainText("messages/authenticate_primary_login_link.txt")
	TemplateMessageAuthenticatePrimaryLoginLinkEmailHTML = template.RegisterMessageHTML("messages/authenticate_primary_login_link.html")

	TemplateMessageAuthenticateSecondaryOOBSMSTXT    = template.RegisterMessagePlainText("messages/authenticate_secondary_oob_sms.txt")
	TemplateMessageAuthenticateSecondaryOOBEmailTXT  = template.RegisterMessagePlainText("messages/authenticate_secondary_oob_email.txt")
	TemplateMessageAuthenticateSecondaryOOBEmailHTML = template.RegisterMessageHTML("messages/authenticate_secondary_oob_email.html")

	TemplateMessageSetupPrimaryLoginLinkEmailTXT  = template.RegisterMessagePlainText("messages/setup_primary_login_link.txt")
	TemplateMessageSetupPrimaryLoginLinkEmailHTML = template.RegisterMessageHTML("messages/setup_primary_login_link.html")

	TemplateMessageSetupSecondaryLoginLinkEmailTXT  = template.RegisterMessagePlainText("messages/setup_secondary_login_link.txt")
	TemplateMessageSetupSecondaryLoginLinkEmailHTML = template.RegisterMessageHTML("messages/setup_secondary_login_link.html")

	TemplateMessageAuthenticateSecondaryLoginLinkEmailTXT  = template.RegisterMessagePlainText("messages/authenticate_secondary_login_link.txt")
	TemplateMessageAuthenticateSecondaryLoginLinkEmailHTML = template.RegisterMessageHTML("messages/authenticate_secondary_login_link.html")

	TemplateMessageForgotPasswordLinkSMSTXT    = template.RegisterMessagePlainText("messages/forgot_password_sms.txt")
	TemplateMessageForgotPasswordLinkEmailTXT  = template.RegisterMessagePlainText("messages/forgot_password_email.txt")
	TemplateMessageForgotPasswordLinkEmailHTML = template.RegisterMessageHTML("messages/forgot_password_email.html")

	TemplateMessageForgotPasswordOOBSMSTXT    = template.RegisterMessagePlainText("messages/forgot_password_oob_sms.txt")
	TemplateMessageForgotPasswordOOBEmailTXT  = template.RegisterMessagePlainText("messages/forgot_password_oob_email.txt")
	TemplateMessageForgotPasswordOOBEmailHTML = template.RegisterMessageHTML("messages/forgot_password_oob_email.html")

	TemplateWhatsappOTPCodeTXT = template.RegisterMessagePlainText("messages/whatsapp_otp_code.txt")
)

var (
	messageVerification = &translation.MessageSpec{
		Name:              "verification",
		TXTEmailTemplate:  TemplateMessageVerificationEmailTXT,
		HTMLEmailTemplate: TemplateMessageVerificationEmailHTML,
		SMSTemplate:       TemplateMessageVerificationSMSTXT,
	}
	messageSetupPrimaryOOB = &translation.MessageSpec{
		Name:              "setup-primary-oob",
		TXTEmailTemplate:  TemplateMessageSetupPrimaryOOBEmailTXT,
		HTMLEmailTemplate: TemplateMessageSetupPrimaryOOBEmailHTML,
		SMSTemplate:       TemplateMessageSetupPrimaryOOBSMSTXT,
	}
	messageSetupPrimaryLoginLink = &translation.MessageSpec{
		Name:              "setup-primary-login-link",
		TXTEmailTemplate:  TemplateMessageSetupPrimaryLoginLinkEmailTXT,
		HTMLEmailTemplate: TemplateMessageSetupPrimaryLoginLinkEmailHTML,
	}
	messageSetupSecondaryOOB = &translation.MessageSpec{
		Name:              "setup-secondary-oob",
		TXTEmailTemplate:  TemplateMessageSetupSecondaryOOBEmailTXT,
		HTMLEmailTemplate: TemplateMessageSetupSecondaryOOBEmailHTML,
		SMSTemplate:       TemplateMessageSetupSecondaryOOBSMSTXT,
	}
	messageSetupSecondaryLoginLink = &translation.MessageSpec{
		Name:              "setup-secondary-login-link",
		TXTEmailTemplate:  TemplateMessageSetupSecondaryLoginLinkEmailTXT,
		HTMLEmailTemplate: TemplateMessageSetupSecondaryLoginLinkEmailHTML,
	}
	messageAuthenticatePrimaryOOB = &translation.MessageSpec{
		Name:              "authenticate-primary-oob",
		TXTEmailTemplate:  TemplateMessageAuthenticatePrimaryOOBEmailTXT,
		HTMLEmailTemplate: TemplateMessageAuthenticatePrimaryOOBEmailHTML,
		SMSTemplate:       TemplateMessageAuthenticatePrimaryOOBSMSTXT,
	}
	messageAuthenticatePrimaryLoginLink = &translation.MessageSpec{
		Name:              "authenticate-primary-login-link",
		TXTEmailTemplate:  TemplateMessageAuthenticatePrimaryLoginLinkEmailTXT,
		HTMLEmailTemplate: TemplateMessageAuthenticatePrimaryLoginLinkEmailHTML,
	}
	messageAuthenticateSecondaryOOB = &translation.MessageSpec{
		Name:              "authenticate-secondary-oob",
		TXTEmailTemplate:  TemplateMessageAuthenticateSecondaryOOBEmailTXT,
		HTMLEmailTemplate: TemplateMessageAuthenticateSecondaryOOBEmailHTML,
		SMSTemplate:       TemplateMessageAuthenticateSecondaryOOBSMSTXT,
	}
	messageAuthenticateSecondaryLoginLink = &translation.MessageSpec{
		Name:              "authenticate-secondary-login-link",
		TXTEmailTemplate:  TemplateMessageAuthenticateSecondaryLoginLinkEmailTXT,
		HTMLEmailTemplate: TemplateMessageAuthenticateSecondaryLoginLinkEmailHTML,
	}
	messageForgotPasswordLink = &translation.MessageSpec{
		Name:              "forgot-password",
		TXTEmailTemplate:  TemplateMessageForgotPasswordLinkEmailTXT,
		HTMLEmailTemplate: TemplateMessageForgotPasswordLinkEmailHTML,
		SMSTemplate:       TemplateMessageForgotPasswordLinkSMSTXT,
	}
	messageForgotPasswordOOB = &translation.MessageSpec{
		Name:              "forgot-password-oob",
		TXTEmailTemplate:  TemplateMessageForgotPasswordOOBEmailTXT,
		HTMLEmailTemplate: TemplateMessageForgotPasswordOOBEmailHTML,
		SMSTemplate:       TemplateMessageForgotPasswordOOBSMSTXT,
	}
	messageWhatsappCode = &translation.MessageSpec{
		Name:             "whatsapp-code",
		WhatsappTemplate: TemplateWhatsappOTPCodeTXT,
	}
)
