package authenticator

type MagicLinkOTPSpec struct {
	Email string `json:"email,omitempty"`
	Code  string `json:"code,omitempty"`
}
