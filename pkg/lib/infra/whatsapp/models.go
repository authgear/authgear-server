package whatsapp

type SendAuthenticationOTPOptions struct {
	To  string
	OTP string
}

type WhatsappAPIErrorResponse struct {
	Errors []WhatsappAPIErrorDetail `json:"errors,omitempty"`
}

func (r *WhatsappAPIErrorResponse) FirstErrorCode() (int, bool) {
	if r.Errors != nil && len(r.Errors) > 0 {
		return (r.Errors)[0].Code, true
	}
	return -1, false
}

type WhatsappAPIErrorDetail struct {
	Code    int    `json:"code"`
	Title   string `json:"title"`
	Details string `json:"details"`
}

const (
	errorCodeInvalidUser = 1013
)
