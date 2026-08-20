package authenticationflow

import (
	"context"
	"errors"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/botprotection"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

func TestErrorBotProtectionVerificationCause(t *testing.T) {
	Convey("the cause of ErrorBotProtectionVerification", t, func() {
		ctx := context.Background()
		cause := errors.Join(botprotection.ErrVerificationServiceUnavailable, fmt.Errorf("internal-error"))
		errBPV := NewErrorBotProtectionVerificationServiceUnavailable(cause)

		// This is what accept() returns.
		err := errorutil.WithSecondaryError(botprotection.ErrVerificationServiceUnavailable, errBPV.Cause)

		Convey("it is visible in the log", func() {
			So(errorutil.Summary(err), ShouldContainSubstring, "internal-error")
		})

		Convey("it is invisible to the end user", func() {
			apiError := apierrors.AsAPIErrorWithContext(ctx, err)
			So(apiError.Reason, ShouldEqual, "BotProtectionVerificationServiceUnavailable")
			So(apiError.Code, ShouldEqual, 503)
			So(apiError.Message, ShouldEqual, "bot protection service unavailable")
			So(apiError.Info_ReadOnly, ShouldBeEmpty)
		})

		Convey("it does not affect errors.Is", func() {
			So(errors.Is(err, botprotection.ErrVerificationServiceUnavailable), ShouldBeTrue)
			So(errors.Is(err, botprotection.ErrVerificationFailed), ShouldBeFalse)
		})

		Convey("it is optional", func() {
			err := errorutil.WithSecondaryError(botprotection.ErrVerificationServiceUnavailable, ErrorBotProtectionVerificationServiceUnavailable.Cause)
			So(errors.Is(err, botprotection.ErrVerificationServiceUnavailable), ShouldBeTrue)
			So(ErrorBotProtectionVerificationServiceUnavailable.Error(), ShouldEqual, "bot protection verification status: service-unavailable")
		})
	})
}
