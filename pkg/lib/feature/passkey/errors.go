package passkey

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrUserNotFound = apierrors.NotFound.WithReason("UserNotFound").New("user not found")
