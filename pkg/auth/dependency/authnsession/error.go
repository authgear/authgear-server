package authnsession

import (
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

var (
	InvalidAuthenticationSession  skyerr.Kind = skyerr.Invalid.WithReason("InvalidAuthenticationSession")
	AuthenticationSessionRequired skyerr.Kind = skyerr.Unauthorized.WithReason("AuthenticationSession")
)

var errInvalidToken = InvalidAuthenticationSession.New("invalid authentication session token")
