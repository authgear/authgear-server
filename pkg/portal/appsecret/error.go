package appsecret

import "github.com/authgear/authgear-server/pkg/api/apierrors"

var ErrTokenNotFound = apierrors.NotFound.WithReason("TokenNotFound").New("token not found")
