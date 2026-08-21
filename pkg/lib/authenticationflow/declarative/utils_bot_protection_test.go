package declarative

import (
	"context"
	"errors"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/botprotection"
)

type stubBotProtectionService struct {
	Err error
}

func (s *stubBotProtectionService) Verify(ctx context.Context, response string) error {
	return s.Err
}

func TestVerifyBotProtectionToken(t *testing.T) {
	Convey("verifyBotProtectionToken", t, func() {
		ctx := context.Background()

		test := func(verifyErr error) (*authflow.ErrorBotProtectionVerification, error) {
			deps := &authflow.Dependencies{
				BotProtection: &stubBotProtectionService{Err: verifyErr},
			}
			bpSpecialErr, err := verifyBotProtectionToken(ctx, deps, "token")
			var errBPV *authflow.ErrorBotProtectionVerification
			if errors.As(bpSpecialErr, &errBPV) {
				return errBPV, err
			}
			return nil, err
		}

		Convey("it keeps the cause of service unavailable", func() {
			cause := fmt.Errorf("connection reset by peer")
			verifyErr := errors.Join(botprotection.ErrVerificationServiceUnavailable, cause)

			errBPV, err := test(verifyErr)
			So(err, ShouldBeNil)
			So(errBPV, ShouldNotBeNil)
			So(errBPV.Status, ShouldEqual, authflow.ErrorBotProtectionVerificationStatusServiceUnavailable)
			So(errors.Is(errBPV.Cause, cause), ShouldBeTrue)
			So(errBPV.Error(), ShouldContainSubstring, "connection reset by peer")
		})

		Convey("it keeps the cause of failed", func() {
			cause := fmt.Errorf("timeout-or-duplicate")
			verifyErr := errors.Join(botprotection.ErrVerificationFailed, cause)

			errBPV, err := test(verifyErr)
			So(err, ShouldBeNil)
			So(errBPV, ShouldNotBeNil)
			So(errBPV.Status, ShouldEqual, authflow.ErrorBotProtectionVerificationStatusFailed)
			So(errors.Is(errBPV.Cause, cause), ShouldBeTrue)
			So(errBPV.Error(), ShouldContainSubstring, "timeout-or-duplicate")
		})

		Convey("it returns success without cause", func() {
			errBPV, err := test(nil)
			So(err, ShouldBeNil)
			So(errBPV, ShouldNotBeNil)
			So(errBPV.Status, ShouldEqual, authflow.ErrorBotProtectionVerificationStatusSuccess)
			So(errBPV.Cause, ShouldBeNil)
		})

		Convey("it returns unexpected error as-is", func() {
			unexpectedErr := fmt.Errorf("failed to unmarshal JSON")

			errBPV, err := test(unexpectedErr)
			So(errBPV, ShouldBeNil)
			So(errors.Is(err, unexpectedErr), ShouldBeTrue)
		})
	})
}
