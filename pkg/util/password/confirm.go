package password

import (
	"crypto/subtle"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrNewPasswordTypo = apierrors.Invalid.WithReason("NewPasswordTypo").New("new password typo")

func ConfirmPassword(newPassword, confirmPassword string) error {
	typo := subtle.ConstantTimeCompare([]byte(newPassword), []byte(confirmPassword)) == 0
	if typo {
		return ErrNewPasswordTypo
	}
	return nil
}
