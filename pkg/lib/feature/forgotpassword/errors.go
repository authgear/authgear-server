package forgotpassword

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var PasswordResetFailed = apierrors.Invalid.WithReason("PasswordResetFailed")

var ErrInvalidCode = PasswordResetFailed.NewWithCause("invalid code", apierrors.StringCause("InvalidCode"))
var ErrUsedCode = PasswordResetFailed.NewWithCause("used code", apierrors.StringCause("UsedCode"))

var SendCodeFailed = apierrors.Invalid.WithReason("ForgotPasswordFailed")

var ErrFeatureDisabled = SendCodeFailed.NewWithCause("forgot password is disabled", apierrors.StringCause("FeatureDisabled"))
var ErrUserNotFound = SendCodeFailed.NewWithCause("specified user not found", apierrors.StringCause("UserNotFound"))

var ErrEmailIdentityNotFound = apierrors.Invalid.WithReason("EmailIdentityNotFound").NewWithCause("email identity not found", apierrors.StringCause("EmailIdentityNotFound"))
