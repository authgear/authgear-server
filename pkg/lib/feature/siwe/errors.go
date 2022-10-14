package siwe

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var InvalidNonce = apierrors.Forbidden.WithReason("InvalidNonce")
var InvalidNetwork = apierrors.BadRequest.WithReason("InvalidNetwork")

var ErrNonceNotFound = InvalidNonce.NewWithCause("nonce is expired or invalid", apierrors.StringCause("NonceNotFound"))
