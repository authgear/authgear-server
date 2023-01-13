package webapp

type FlashMessageType string

const (
	FlashMessageTypeResendCodeSuccess      FlashMessageType = "resend_code_success"
	FlashMessageTypeResendMagicLinkSuccess FlashMessageType = "resend_magic_link_success"
)
