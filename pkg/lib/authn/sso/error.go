package sso

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var InvalidConfiguration = apierrors.InternalError.WithReason("InvalidConfiguration")
var OAuthProtocolError = apierrors.BadRequest.WithReason("OAuthProtocolError")
