package translation

import "github.com/authgear/authgear-server/pkg/util/template"

type MessageType string

const (
	MessageTypeVerification               MessageType = "verification"
	MessageTypeSetupPrimaryOOB            MessageType = "setup-primary-oob"
	MessageTypeSetupSecondaryOOB          MessageType = "setup-secondary-oob"
	MessageTypeAuthenticatePrimaryOOB     MessageType = "authenticate-primary-oob"
	MessageTypeAuthenticateSecondaryOOB   MessageType = "authenticate-secondary-oob"
	MessageTypeForgotPassword             MessageType = "forgot-password"
	MessageTypeSendPasswordToExistingUser MessageType = "send-password-to-existing-user"
	MessageTypeSendPasswordToNewUser      MessageType = "send-password-to-new-user"
	MessageTypeWhatsappCode               MessageType = "whatsapp-code"
)

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

	TemplateMessageSendPasswordToExistingUserTXT       = template.RegisterMessagePlainText("messages/send_password_to_existing_user_email.txt")
	TemplateMessageSendPasswordToExistingUserEmailHTML = template.RegisterMessageHTML("messages/send_password_to_existing_user_email.html")

	TemplateMessageSendPasswordToNewUserTXT       = template.RegisterMessagePlainText("messages/send_password_to_new_user_email.txt")
	TemplateMessageSendPasswordToNewUserEmailHTML = template.RegisterMessageHTML("messages/send_password_to_new_user_email.html")
)

var (
	MessageVerification = &MessageSpec{
		Name:              "verification",
		TXTEmailTemplate:  TemplateMessageVerificationEmailTXT,
		HTMLEmailTemplate: TemplateMessageVerificationEmailHTML,
		SMSTemplate:       TemplateMessageVerificationSMSTXT,
	}
	MessageSetupPrimaryOOB = &MessageSpec{
		Name:              "setup-primary-oob",
		TXTEmailTemplate:  TemplateMessageSetupPrimaryOOBEmailTXT,
		HTMLEmailTemplate: TemplateMessageSetupPrimaryOOBEmailHTML,
		SMSTemplate:       TemplateMessageSetupPrimaryOOBSMSTXT,
	}
	MessageSetupPrimaryLoginLink = &MessageSpec{
		Name:              "setup-primary-login-link",
		TXTEmailTemplate:  TemplateMessageSetupPrimaryLoginLinkEmailTXT,
		HTMLEmailTemplate: TemplateMessageSetupPrimaryLoginLinkEmailHTML,
	}
	MessageSetupSecondaryOOB = &MessageSpec{
		Name:              "setup-secondary-oob",
		TXTEmailTemplate:  TemplateMessageSetupSecondaryOOBEmailTXT,
		HTMLEmailTemplate: TemplateMessageSetupSecondaryOOBEmailHTML,
		SMSTemplate:       TemplateMessageSetupSecondaryOOBSMSTXT,
	}
	MessageSetupSecondaryLoginLink = &MessageSpec{
		Name:              "setup-secondary-login-link",
		TXTEmailTemplate:  TemplateMessageSetupSecondaryLoginLinkEmailTXT,
		HTMLEmailTemplate: TemplateMessageSetupSecondaryLoginLinkEmailHTML,
	}
	MessageAuthenticatePrimaryOOB = &MessageSpec{
		Name:              "authenticate-primary-oob",
		TXTEmailTemplate:  TemplateMessageAuthenticatePrimaryOOBEmailTXT,
		HTMLEmailTemplate: TemplateMessageAuthenticatePrimaryOOBEmailHTML,
		SMSTemplate:       TemplateMessageAuthenticatePrimaryOOBSMSTXT,
	}
	MessageAuthenticatePrimaryLoginLink = &MessageSpec{
		Name:              "authenticate-primary-login-link",
		TXTEmailTemplate:  TemplateMessageAuthenticatePrimaryLoginLinkEmailTXT,
		HTMLEmailTemplate: TemplateMessageAuthenticatePrimaryLoginLinkEmailHTML,
	}
	MessageAuthenticateSecondaryOOB = &MessageSpec{
		Name:              "authenticate-secondary-oob",
		TXTEmailTemplate:  TemplateMessageAuthenticateSecondaryOOBEmailTXT,
		HTMLEmailTemplate: TemplateMessageAuthenticateSecondaryOOBEmailHTML,
		SMSTemplate:       TemplateMessageAuthenticateSecondaryOOBSMSTXT,
	}
	MessageAuthenticateSecondaryLoginLink = &MessageSpec{
		Name:              "authenticate-secondary-login-link",
		TXTEmailTemplate:  TemplateMessageAuthenticateSecondaryLoginLinkEmailTXT,
		HTMLEmailTemplate: TemplateMessageAuthenticateSecondaryLoginLinkEmailHTML,
	}
	MessageForgotPasswordLink = &MessageSpec{
		Name:              "forgot-password",
		TXTEmailTemplate:  TemplateMessageForgotPasswordLinkEmailTXT,
		HTMLEmailTemplate: TemplateMessageForgotPasswordLinkEmailHTML,
		SMSTemplate:       TemplateMessageForgotPasswordLinkSMSTXT,
	}
	MessageForgotPasswordOOB = &MessageSpec{
		Name:              "forgot-password-oob",
		TXTEmailTemplate:  TemplateMessageForgotPasswordOOBEmailTXT,
		HTMLEmailTemplate: TemplateMessageForgotPasswordOOBEmailHTML,
		SMSTemplate:       TemplateMessageForgotPasswordOOBSMSTXT,
	}
	MessageWhatsappCode = &MessageSpec{
		Name:             "whatsapp-code",
		WhatsappTemplate: TemplateWhatsappOTPCodeTXT,
	}
	MessageSendPasswordToExistingUser = &MessageSpec{
		Name:              "send-password-to-existing-user",
		TXTEmailTemplate:  TemplateMessageSendPasswordToExistingUserTXT,
		HTMLEmailTemplate: TemplateMessageSendPasswordToExistingUserEmailHTML,
	}
	MessageSendPasswordToNewUser = &MessageSpec{
		Name:              "send-password-to-new-user",
		TXTEmailTemplate:  TemplateMessageSendPasswordToNewUserTXT,
		HTMLEmailTemplate: TemplateMessageSendPasswordToNewUserEmailHTML,
	}
)
