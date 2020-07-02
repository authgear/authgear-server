package webapp

import (
	"github.com/authgear/authgear-server/pkg/core/skyerr"
)

var ErrOAuthProviderNotFound = skyerr.NotFound.WithReason("OAuthProviderNotFound").New("oauth provider not found")
