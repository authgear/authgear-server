package forgotpassword

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var PasswordResetFailed = apierrors.Invalid.WithReason("PasswordResetFailed")

var ErrInvalidCode = PasswordResetFailed.NewWithCause("invalid code", apierrors.StringCause("InvalidCode"))
var ErrUsedCode = PasswordResetFailed.NewWithCause("used code", apierrors.StringCause("UsedCode"))
var ErrExpiredCode = PasswordResetFailed.NewWithCause("expired code", apierrors.StringCause("ExpiredCode"))
var ErrNoPassword = PasswordResetFailed.NewWithCause("specified user cannot be logged in using password", apierrors.StringCause("NoPassword"))

var SendCodeFailed = apierrors.Invalid.WithReason("ForgotPasswordFailed")

var ErrUserNotFound = SendCodeFailed.NewWithCause("specified user not found", apierrors.StringCause("UserNotFound"))
