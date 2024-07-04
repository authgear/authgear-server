package facade

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrUserIsAnonymized = apierrors.Invalid.WithReason("UserIsAnonymized").New("user is anonymized")
