package declarative

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"

	. "github.com/smartystreets/goconvey/convey"
)

type captureOTPCodeServiceForVerifyClaim struct {
	lastGenerateOpts *otp.GenerateOptions
	lastInspectOpts  *otp.InspectStateOptions
}

func (s *captureOTPCodeServiceForVerifyClaim) GenerateOTP(ctx context.Context, kind otp.Kind, target string, form otp.Form, opt *otp.GenerateOptions) (string, error) {
	s.lastGenerateOpts = opt
	return "111111", nil
}

func (s *captureOTPCodeServiceForVerifyClaim) VerifyOTP(ctx context.Context, kind otp.Kind, target string, code string, opts *otp.VerifyOptions) error {
	return nil
}

func (s *captureOTPCodeServiceForVerifyClaim) InspectState(ctx context.Context, kind otp.Kind, target string, opts *otp.InspectStateOptions) (*otp.State, error) {
	s.lastInspectOpts = opts
	return &otp.State{CanResendAt: time.Unix(1700000000, 0).UTC()}, nil
}

func (s *captureOTPCodeServiceForVerifyClaim) LookupCode(ctx context.Context, purpose otp.Purpose, code string) (string, error) {
	return "", nil
}

func (s *captureOTPCodeServiceForVerifyClaim) SetSubmittedCode(ctx context.Context, kind otp.Kind, target string, code string) (*otp.State, error) {
	return nil, nil
}

func TestNodeVerifyClaimPassesAuthenticationFlowID(t *testing.T) {
	Convey("NodeVerifyClaim passes authflow FlowID to OTP generate and inspect calls", t, func() {
		otpCodes := &captureOTPCodeServiceForVerifyClaim{}
		deps := &authflow.Dependencies{
			Config:      &config.AppConfig{},
			OTPCodes:    otpCodes,
			HTTPRequest: &http.Request{Header: http.Header{}},
		}
		session := &authflow.Session{
			FlowID:    "flow-1",
			UILocales: "en",
		}
		ctx := session.MakeContext(context.Background(), deps)
		node := &NodeVerifyClaim{
			UserID:     "user-1",
			Purpose:    otp.PurposeVerification,
			Form:       otp.FormCode,
			ClaimName:  model.ClaimPhoneNumber,
			ClaimValue: "+85265000001",
			Channel:    model.AuthenticatorOOBChannelSMS,
		}

		_, err := node.GenerateCode(ctx, deps)
		So(err, ShouldBeNil)
		So(otpCodes.lastGenerateOpts, ShouldNotBeNil)
		So(otpCodes.lastGenerateOpts.AuthenticationFlowID, ShouldEqual, "flow-1")

		_, err = node.OutputData(ctx, deps, authflow.Flows{})
		So(err, ShouldBeNil)
		So(otpCodes.lastInspectOpts, ShouldNotBeNil)
		So(otpCodes.lastInspectOpts.AuthenticationFlowID, ShouldEqual, "flow-1")
	})
}
