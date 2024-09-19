package translation

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
