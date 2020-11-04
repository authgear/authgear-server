package webapp

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrInvalidSession = apierrors.Invalid.WithReason("WebUIInvalidSession").New("session expired or invalid")
var ErrSessionNotFound = apierrors.Invalid.WithReason("WebUIInvalidSession").New("session not found")
