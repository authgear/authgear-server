package webapp

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var WebUIInvalidSession = apierrors.Invalid.WithReason("WebUIInvalidSession")
var WebUISessionCompleted = apierrors.Invalid.WithReason("WebUISessionCompleted")

var ErrInvalidSession = WebUIInvalidSession.New("session expired or invalid")
var ErrSessionNotFound = WebUIInvalidSession.New("session not found")
var ErrSessionStepMismatch = WebUIInvalidSession.New("session step does match request path")
var ErrSessionCompleted = WebUISessionCompleted.New("session completed")

func init() {
	apierrors.SkipLoggingForKinds[WebUISessionCompleted] = true
}
