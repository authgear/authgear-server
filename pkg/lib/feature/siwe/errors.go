package siwe

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var InvalidNonce = apierrors.Forbidden.WithReason("InvalidNonce")
var InvalidNetwork = apierrors.BadRequest.WithReason("InvalidNetwork")

var ErrNonceNotFound = InvalidNonce.NewWithCause("nonce is expired or invalid", apierrors.StringCause("NonceNotFound"))
var ErrMismatchNetwork = InvalidNetwork.NewWithCause("network does not match expected network", apierrors.StringCause("MismatchNetwork"))
