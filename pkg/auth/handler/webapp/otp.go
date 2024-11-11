package webapp

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
)

type OTPCodeService interface {
	VerifyOTP(ctx context.Context, kind otp.Kind, target string, otp string, opts *otp.VerifyOptions) error
	InspectState(ctx context.Context, kind otp.Kind, target string) (*otp.State, error)

	LookupCode(ctx context.Context, purpose otp.Purpose, code string) (target string, err error)
	SetSubmittedCode(ctx context.Context, kind otp.Kind, target string, code string) (*otp.State, error)
}
