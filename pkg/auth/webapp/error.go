package webapp

import (
	"github.com/authgear/authgear-server/pkg/lib/api/apierrors"
)

var ErrOAuthProviderNotFound = apierrors.NotFound.WithReason("OAuthProviderNotFound").New("oauth provider not found")
