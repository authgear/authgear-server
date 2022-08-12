package passkey

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrUserNotFound = apierrors.NotFound.WithReason("UserNotFound").New("user not found")
var ErrSessionNotFound = apierrors.NotFound.WithReason("WebAuthnSessionNotFound").New("webauthn session not found")
