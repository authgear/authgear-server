package handler

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/oauth"
)

func IsConsentRequiredError(err error) bool {
	return errors.Is(err, oauth.ErrAuthorizationScopesNotGranted) || errors.Is(err, oauth.ErrAuthorizationNotFound)
}
