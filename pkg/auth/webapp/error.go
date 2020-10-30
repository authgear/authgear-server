package webapp

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrOAuthProviderNotFound = apierrors.NotFound.WithReason("OAuthProviderNotFound").New("oauth provider not found")
var ErrInvalidState = apierrors.Invalid.WithReason("WebUIInvalidState").New("the claimed session is invalid")
var ErrInvalidUserAgentToken = apierrors.Invalid.WithReason("WebUIInvalidUserAgentToken").New("invalid user agent token")
