package authenticator

type MagicLinkOTPSpec struct {
	Email string `json:"email,omitempty"`
	Token string `json:"token,omitempty"`
}
