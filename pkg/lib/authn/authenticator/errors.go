package authenticator

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/authn"
)

var ErrAuthenticatorNotFound = errors.New("authenticator not found")

var ErrInvalidCredentials = errors.New("invalid credentials")

func NewErrDuplicatedAuthenticator(typ authn.AuthenticatorType) error {
	return apierrors.Invalid.WithReason("InvariantViolated").
		NewWithCause(
			"duplicated authenticator",
			apierrors.MapCause{
				CauseKind: "DuplicatedAuthenticator",
				Data: map[string]interface{}{
					"AuthenticatorType": string(typ),
				},
			},
		)
}
