package oauthrelyingpartyutil

import (
	"net/url"
	"strings"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"
)

func GetCode(query string) (code string, err error) {
	// query may start with a ?, remove it.
	query = strings.TrimPrefix(query, "?")
	form, err := url.ParseQuery(query)
	if err != nil {
		return
	}

	error := form.Get("error")
	errorDescription := form.Get("error_description")
	errorURI := form.Get("error_uri")
	if error != "" {
		err = &oauthrelyingparty.ErrorResponse{
			Error_:           error,
			ErrorDescription: errorDescription,
			ErrorURI:         errorURI,
		}
		return
	}

	code = form.Get("code")
	return
}
