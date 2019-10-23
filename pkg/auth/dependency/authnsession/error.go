package authnsession

import (
	skyerr "github.com/skygeario/skygear-server/pkg/core/xskyerr"
)

var (
	InvalidAuthenticationSession  skyerr.Kind = skyerr.Invalid.WithReason("InvalidAuthenticationSession")
	AuthenticationSessionRequired skyerr.Kind = skyerr.Unauthorized.WithReason("AuthenticationSession")
)

var errInvalidToken = InvalidAuthenticationSession.New("invalid authentication session token")
