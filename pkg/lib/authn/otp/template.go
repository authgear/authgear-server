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
)

type messageTemplateContext struct {
	Email string
	Phone string
	Code  string
	URL   string
	Host  string
}

var (
	TemplateMessageVerificationSMSTXT    = template.RegisterPlainText("messages/verification_sms.txt")
	TemplateMessageVerificationEmailTXT  = template.RegisterPlainText("messages/verification_email.txt")
	TemplateMessageVerificationEmailHTML = template.RegisterHTML("messages/verification_email.html")

	TemplateMessageSetupPrimaryOOBSMSTXT    = template.RegisterPlainText("messages/setup_primary_oob_sms.txt")
	TemplateMessageSetupPrimaryOOBEmailTXT  = template.RegisterPlainText("messages/setup_primary_oob_email.txt")
	TemplateMessageSetupPrimaryOOBEmailHTML = template.RegisterHTML("messages/setup_primary_oob_email.html")

	TemplateMessageSetupSecondaryOOBSMSTXT    = template.RegisterPlainText("messages/setup_secondary_oob_sms.txt")
	TemplateMessageSetupSecondaryOOBEmailTXT  = template.RegisterPlainText("messages/setup_secondary_oob_email.txt")
	TemplateMessageSetupSecondaryOOBEmailHTML = template.RegisterHTML("messages/setup_secondary_oob_email.txt")

	TemplateMessageAuthenticatePrimaryOOBSMSTXT    = template.RegisterPlainText("messages/authenticate_primary_oob_sms.txt")
	TemplateMessageAuthenticatePrimaryOOBEmailTXT  = template.RegisterPlainText("messages/authenticate_primary_oob_email.txt")
	TemplateMessageAuthenticatePrimaryOOBEmailHTML = template.RegisterHTML("messages/authenticate_primary_oob_email.html")

	TemplateMessageAuthenticatePrimaryLoginLinkEmailTXT  = template.RegisterPlainText("messages/authenticate_primary_login_link.txt")
	TemplateMessageAuthenticatePrimaryLoginLinkEmailHTML = template.RegisterHTML("messages/authenticate_primary_login_link.html")

	TemplateMessageAuthenticateSecondaryOOBSMSTXT    = template.RegisterPlainText("messages/authenticate_secondary_oob_sms.txt")
	TemplateMessageAuthenticateSecondaryOOBEmailTXT  = template.RegisterPlainText("messages/authenticate_secondary_oob_email.txt")
	TemplateMessageAuthenticateSecondaryOOBEmailHTML = template.RegisterHTML("messages/authenticate_secondary_oob_email.html")

	TemplateMessageSetupPrimaryLoginLinkEmailTXT  = template.RegisterPlainText("messages/setup_primary_login_link.txt")
	TemplateMessageSetupPrimaryLoginLinkEmailHTML = template.RegisterHTML("messages/setup_primary_login_link.html")

	TemplateMessageSetupSecondaryLoginLinkEmailTXT  = template.RegisterPlainText("messages/setup_secondary_login_link.txt")
	TemplateMessageSetupSecondaryLoginLinkEmailHTML = template.RegisterHTML("messages/setup_secondary_login_link.html")

	TemplateMessageAuthenticateSecondaryLoginLinkEmailTXT  = template.RegisterPlainText("messages/authenticate_secondary_login_link.txt")
	TemplateMessageAuthenticateSecondaryLoginLinkEmailHTML = template.RegisterHTML("messages/authenticate_secondary_login_link.html")
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
)
