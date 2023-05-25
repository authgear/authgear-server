package webapp

import "github.com/authgear/authgear-server/pkg/lib/authn/otp"

type OTPCodeService interface {
	VerifyOTP(kind otp.Kind, target string, otp string, opts *otp.VerifyOptions) error
	InspectState(kind otp.Kind, target string) (*otp.State, error)

	LookupCode(purpose otp.Purpose, code string) (target string, err error)
	SetSubmittedCode(kind otp.Kind, target string, code string) (*otp.State, error)
}
