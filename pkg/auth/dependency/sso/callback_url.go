package sso

import (
	"github.com/skygeario/skygear-server/pkg/core/errors"
)

func ValidateCallbackURL(allowedCallbackURLs []string, callbackURL string) (err error) {
	// The logic of this function must be in sync with the inline javascript implementation.
	if callbackURL == "" {
		err = errors.New("missing callback URL")
		return
	}

	for _, v := range allowedCallbackURLs {
		if callbackURL == v {
			return nil
		}
	}

	err = errors.New("callback URL is not whitelisted")
	return
}
