package declarative

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
)

type VerifyOOBOTPData struct {
	TypedData
	Channel                        model.AuthenticatorOOBChannel `json:"channel,omitempty"`
	OTPForm                        otp.Form                      `json:"otp_form,omitempty"`
	WebsocketURL                   string                        `json:"websocket_url,omitempty"`
	MaskedClaimValue               string                        `json:"masked_claim_value,omitempty"`
	CodeLength                     int                           `json:"code_length,omitempty"`
	CanResendAt                    time.Time                     `json:"can_resend_at,omitempty"`
	CanCheck                       bool                          `json:"can_check"`
	FailedAttemptRateLimitExceeded bool                          `json:"failed_attempt_rate_limit_exceeded"`
	DeliveryStatus                 model.OTPDeliveryStatus       `json:"delivery_status"`
}

func NewVerifyOOBOTPData(d VerifyOOBOTPData) VerifyOOBOTPData {
	d.Type = DataTypeVerifyOOBOTPData
	return d
}

var _ authflow.Data = VerifyOOBOTPData{}

func (m VerifyOOBOTPData) Data() {}
