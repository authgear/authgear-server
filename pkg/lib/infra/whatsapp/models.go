package whatsapp

type SendAuthenticationOTPOptions struct {
	To  string
	OTP string
}

type WhatsappOnPremisesAPIErrorResponse struct {
	Errors []WhatsappOnPremisesAPIErrorDetail `json:"errors,omitempty"`
}

func (r *WhatsappOnPremisesAPIErrorResponse) FirstErrorCode() (int, bool) {
	if r.Errors != nil && len(r.Errors) > 0 {
		return (r.Errors)[0].Code, true
	}
	return -1, false
}

type WhatsappOnPremisesAPIErrorDetail struct {
	Code    int    `json:"code"`
	Title   string `json:"title"`
	Details string `json:"details"`
}

// The documented Meta Graph API error fields are
// https://developers.facebook.com/docs/graph-api/guides/error-handling
//
// But the observed Cloud API error actually looks like
//
//	{
//	    "error": {
//	        "message": "(#100) Invalid parameter",
//	        "type": "OAuthException",
//	        "code": 100,
//	        "error_data": {
//	            "messaging_product": "whatsapp",
//	            "details": "Parameter Invalid"
//	        },
//	        "fbtrace_id": "AFyaw0muneDI7Xn8cCGcnGG"
//	    }
//	}
//
// So we just deserialize the common fields.
type WhatsappCloudAPIErrorResponseError struct {
	Message   string `json:"message,omitempty"`
	Type      string `json:"type,omitempty"`
	Code      int    `json:"code"`
	FbtraceID string `json:"fbtrace_id"`
}

type WhatsappCloudAPIErrorResponse struct {
	Error *WhatsappCloudAPIErrorResponseError `json:"error,omitempty"`
}

type WhatsappMessageStatus string

const (
	WhatsappMessageStatusAccepted  WhatsappMessageStatus = "accepted"
	WhatsappMessageStatusSent      WhatsappMessageStatus = "sent"
	WhatsappMessageStatusDelivered WhatsappMessageStatus = "delivered"
	WhatsappMessageStatusRead      WhatsappMessageStatus = "read"
	WhatsappMessageStatusFailed    WhatsappMessageStatus = "failed"
)

type WhatsappSendMessageResponse struct {
	MessagingProduct string `json:"messaging_product"`
	Contacts         []struct {
		Input string `json:"input"`
		WaID  string `json:"wa_id"`
	} `json:"contacts"`
	Messages []struct {
		ID            string `json:"id"`
		MessageStatus string `json:"message_status"`
	} `json:"messages"`
}

type WhatsappStatusError struct {
	Code    int    `json:"code"`
	Title   string `json:"title"`
	Message string `json:"message"`
}
