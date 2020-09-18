package password

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrNewPasswordTypo = apierrors.Invalid.WithReason("NewPasswordTypo").New("new password typo")

// ConfirmPassword checks if the user has made a typo mistake when creating their new password.
// Both input are given by the same person, constant time comparison is not needed here.
func ConfirmPassword(newPassword, confirmPassword string) error {
	typo := newPassword != confirmPassword
	if typo {
		return ErrNewPasswordTypo
	}
	return nil
}
