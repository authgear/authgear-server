package declarative

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"

	. "github.com/smartystreets/goconvey/convey"
)

type captureOTPCodeServiceForAuthnOOB struct {
	lastGenerateOpts *otp.GenerateOptions
	lastInspectOpts  *otp.InspectStateOptions
}

func (s *captureOTPCodeServiceForAuthnOOB) GenerateOTP(ctx context.Context, kind otp.Kind, target string, form otp.Form, opt *otp.GenerateOptions) (string, error) {
	s.lastGenerateOpts = opt
	return "111111", nil
}

func (s *captureOTPCodeServiceForAuthnOOB) VerifyOTP(ctx context.Context, kind otp.Kind, target string, code string, opts *otp.VerifyOptions) error {
	return nil
}

func (s *captureOTPCodeServiceForAuthnOOB) InspectState(ctx context.Context, kind otp.Kind, target string, opts *otp.InspectStateOptions) (*otp.State, error) {
	s.lastInspectOpts = opts
	return &otp.State{CanResendAt: time.Unix(1700000000, 0).UTC()}, nil
}

func (s *captureOTPCodeServiceForAuthnOOB) LookupCode(ctx context.Context, purpose otp.Purpose, code string) (string, error) {
	return "", nil
}

func (s *captureOTPCodeServiceForAuthnOOB) SetSubmittedCode(ctx context.Context, kind otp.Kind, target string, code string) (*otp.State, error) {
	return nil, nil
}

func TestNodeAuthenticationOOBPassesAuthenticationFlowID(t *testing.T) {
	Convey("NodeAuthenticationOOB passes authflow FlowID to OTP generate and inspect calls", t, func() {
		otpCodes := &captureOTPCodeServiceForAuthnOOB{}
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
		node := &NodeAuthenticationOOB{
			UserID:  "user-1",
			Purpose: otp.PurposeOOBOTP,
			Form:    otp.FormCode,
			Channel: model.AuthenticatorOOBChannelSMS,
			Info: (&authenticator.OOBOTP{
				UserID:               "user-1",
				Kind:                 string(authenticator.KindPrimary),
				OOBAuthenticatorType: model.AuthenticatorTypeOOBSMS,
				Phone:                "+85265000001",
			}).ToInfo(),
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
