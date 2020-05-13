package webapp

import (
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

var ErrOAuthProviderNotFound = skyerr.NotFound.WithReason("OAuthProviderNotFound").New("oauth provider not found")
