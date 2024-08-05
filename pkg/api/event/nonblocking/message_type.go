package nonblocking

type MessageType string

const (
	MessageTypeVerification             MessageType = "verification"
	MessageTypeSetupPrimaryOOB          MessageType = "setup-primary-oob"
	MessageTypeSetupSecondaryOOB        MessageType = "setup-secondary-oob"
	MessageTypeAuthenticatePrimaryOOB   MessageType = "authenticate-primary-oob"
	MessageTypeAuthenticateSecondaryOOB MessageType = "authenticate-secondary-oob"
	MessageTypeForgotPassword           MessageType = "forgot-password"
	MessageTypeChangePassword           MessageType = "change-password"
	MessageTypeWhatsappCode             MessageType = "whatsapp-code"
)
