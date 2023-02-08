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

type OTPMode string

const (
	OTPModeCode      OTPMode = "code"
	OTPModeMagicLink OTPMode = "magic-link"
)

type MessageTemplateContext struct {
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

	TemplateMessageAuthenticatePrimaryMagicLinkEmailTXT  = template.RegisterPlainText("messages/authenticate_primary_login_link.txt")
	TemplateMessageAuthenticatePrimaryMagicLinkEmailHTML = template.RegisterHTML("messages/authenticate_primary_login_link.html")

	TemplateMessageAuthenticateSecondaryOOBSMSTXT    = template.RegisterPlainText("messages/authenticate_secondary_oob_sms.txt")
	TemplateMessageAuthenticateSecondaryOOBEmailTXT  = template.RegisterPlainText("messages/authenticate_secondary_oob_email.txt")
	TemplateMessageAuthenticateSecondaryOOBEmailHTML = template.RegisterHTML("messages/authenticate_secondary_oob_email.html")

	TemplateMessageSetupPrimaryMagicLinkEmailTXT  = template.RegisterPlainText("messages/setup_primary_login_link.txt")
	TemplateMessageSetupPrimaryMagicLinkEmailHTML = template.RegisterHTML("messages/setup_primary_login_link.html")

	TemplateMessageSetupSecondaryMagicLinkEmailTXT  = template.RegisterPlainText("messages/setup_secondary_login_link.txt")
	TemplateMessageSetupSecondaryMagicLinkEmailHTML = template.RegisterHTML("messages/setup_secondary_login_link.html")

	TemplateMessageAuthenticateSecondaryMagicLinkEmailTXT  = template.RegisterPlainText("messages/authenticate_secondary_login_link.txt")
	TemplateMessageAuthenticateSecondaryMagicLinkEmailHTML = template.RegisterHTML("messages/authenticate_secondary_login_link.html")
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
	messageSetupPrimaryMagicLink = &translation.MessageSpec{
		Name:              "setup-primary-magic-link",
		TXTEmailTemplate:  TemplateMessageSetupPrimaryMagicLinkEmailTXT,
		HTMLEmailTemplate: TemplateMessageSetupPrimaryMagicLinkEmailHTML,
	}
	messageSetupSecondaryOOB = &translation.MessageSpec{
		Name:              "setup-secondary-oob",
		TXTEmailTemplate:  TemplateMessageSetupSecondaryOOBEmailTXT,
		HTMLEmailTemplate: TemplateMessageSetupSecondaryOOBEmailHTML,
		SMSTemplate:       TemplateMessageSetupSecondaryOOBSMSTXT,
	}
	messageSetupSecondaryMagicLink = &translation.MessageSpec{
		Name:              "setup-secondary-magic-link",
		TXTEmailTemplate:  TemplateMessageSetupSecondaryMagicLinkEmailTXT,
		HTMLEmailTemplate: TemplateMessageSetupSecondaryMagicLinkEmailHTML,
	}
	messageAuthenticatePrimaryOOB = &translation.MessageSpec{
		Name:              "authenticate-primary-oob",
		TXTEmailTemplate:  TemplateMessageAuthenticatePrimaryOOBEmailTXT,
		HTMLEmailTemplate: TemplateMessageAuthenticatePrimaryOOBEmailHTML,
		SMSTemplate:       TemplateMessageAuthenticatePrimaryOOBSMSTXT,
	}
	messageAuthenticatePrimaryMagicLink = &translation.MessageSpec{
		Name:              "authenticate-primary-magic-link",
		TXTEmailTemplate:  TemplateMessageAuthenticatePrimaryMagicLinkEmailTXT,
		HTMLEmailTemplate: TemplateMessageAuthenticatePrimaryMagicLinkEmailHTML,
	}
	messageAuthenticateSecondaryOOB = &translation.MessageSpec{
		Name:              "authenticate-secondary-oob",
		TXTEmailTemplate:  TemplateMessageAuthenticateSecondaryOOBEmailTXT,
		HTMLEmailTemplate: TemplateMessageAuthenticateSecondaryOOBEmailHTML,
		SMSTemplate:       TemplateMessageAuthenticateSecondaryOOBSMSTXT,
	}
	messageAuthenticateSecondaryMagicLink = &translation.MessageSpec{
		Name:              "authenticate-secondary-magic-link",
		TXTEmailTemplate:  TemplateMessageAuthenticateSecondaryMagicLinkEmailTXT,
		HTMLEmailTemplate: TemplateMessageAuthenticateSecondaryMagicLinkEmailHTML,
	}
)
