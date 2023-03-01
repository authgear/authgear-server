package webapp

type FlashMessageType string

const (
	FlashMessageTypeResendCodeSuccess      FlashMessageType = "resend_code_success"
	FlashMessageTypeResendLoginLinkSuccess FlashMessageType = "resend_login_link_success"
)
