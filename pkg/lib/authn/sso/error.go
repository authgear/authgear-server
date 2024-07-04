package sso

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var OAuthProtocolError = apierrors.BadRequest.WithReason("OAuthProtocolError")
