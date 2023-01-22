package webapp

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var WebUIInvalidSession = apierrors.Invalid.WithReason("WebUIInvalidSession")

var ErrInvalidSession = WebUIInvalidSession.New("session expired or invalid")
var ErrSessionNotFound = WebUIInvalidSession.New("session not found")
var ErrSessionStepMismatch = WebUIInvalidSession.New("session step does match request path")
