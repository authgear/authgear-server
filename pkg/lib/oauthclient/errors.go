package oauthclient

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrDynamicClientNotFound = apierrors.NotFound.WithReason("DynamicClientNotFound").New("dynamic client not found")
var ErrDynamicClientDuplicateClientID = apierrors.BadRequest.WithReason("DynamicClientDuplicateClientID").New("duplicate dynamic client id")
