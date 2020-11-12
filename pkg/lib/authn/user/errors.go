package user

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrUserNotFound = errors.New("user not found")

var DisabledUser = apierrors.Forbidden.WithReason("DisabledUser")

func NewErrDisabledUser(reason *string) error {
	return DisabledUser.NewWithInfo("user is disabled", map[string]interface{}{
		"reason": reason,
	})
}
