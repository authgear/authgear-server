package authenticator

type OOBOTPSpec struct {
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
	Code  string `json:"code,omitempty"`
}
