package authenticator

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
)

var ErrAuthenticatorNotFound = errors.New("authenticator not found")

func NewErrDuplicatedAuthenticator(typ model.AuthenticatorType) error {
	return apierrors.Invalid.WithReason("DuplicatedAuthenticator").NewWithInfo(
		"duplicated authenticator",
		apierrors.Details{
			"AuthenticatorType": string(typ),
		},
	)
}
