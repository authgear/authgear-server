package sso

import (
	"net/url"
	"strings"

	"github.com/skygeario/skygear-server/pkg/core/errors"
)

// ValidateCallbackURL assumes allowedCallbackURLs are URL without query or fragment.
// It removes query or fragment of callbackURL and check if it appears in allowedCallbackURLs.
// It also ignore trailing slash in allowedCallbackURLs and callbackURL.
func ValidateCallbackURL(allowedCallbackURLs []string, callbackURL string) (err error) {
	// The logic of this function must be in sync with the inline javascript implementation.
	if callbackURL == "" {
		err = errors.New("missing callback URL")
		return
	}

	u, err := url.Parse(callbackURL)
	if err != nil {
		err = errors.New("invalid callback URL")
		return
	}

	u.RawQuery = ""
	u.Fragment = ""
	callbackURL = u.String()

	callbackURL = strings.TrimSuffix(callbackURL, "/")
	for _, v := range allowedCallbackURLs {
		allowed := strings.TrimSuffix(v, "/")
		if callbackURL == allowed {
			return nil
		}
	}

	err = errors.New("callback URL is not whitelisted")
	return
}
