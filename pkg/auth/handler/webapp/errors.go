package webapp

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var UnverifiedIdentityNotFound = apierrors.NotFound.WithReason("UnverifiedIdentityNotFound")

var ErrAuthenticationRequired = apierrors.NewUnauthorized("authentication required")

func NewErrUnverifiedIdentityNotFound(claimName string) error {
	return UnverifiedIdentityNotFound.NewWithInfo("unverified identity not found", map[string]interface{}{
		"claim_name": claimName,
	})
}
