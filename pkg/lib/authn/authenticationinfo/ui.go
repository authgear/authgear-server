package authenticationinfo

import (
	"net/http"
	"net/url"
)

type UIServiceEndpointsProvider interface {
	ConsentEndpointURL() *url.URL
	SAMLLoginFinishURL() *url.URL
}

type UIService struct {
	EndpointsProvider UIServiceEndpointsProvider
}

const queryAuthenticationCode = "code"

func (r *UIService) SetAuthenticationInfoInQuery(redirectURI string, e *Entry) string {
	consentURL := r.EndpointsProvider.ConsentEndpointURL()
	samlLoginFinishURL := r.EndpointsProvider.SAMLLoginFinishURL()

	u, err := url.Parse(redirectURI)
	if err != nil {
		panic(err)
	}

	compareWithSchemeHostPath := func(target *url.URL) bool {
		return u.Scheme == target.Scheme && u.Host == target.Host && u.Path == target.Path
	}

	// When redirectURI is consentURL, it will have client_id, redirect_uri, state in it,
	// so we have to compare them WITHOUT query nor fragment.
	// When we are not redirecting to consentURL or samlLoginFinishURL, we do not set code.
	isConsentURL := compareWithSchemeHostPath(consentURL)
	isSamlLoginFinishURL := compareWithSchemeHostPath(samlLoginFinishURL)
	if !isConsentURL && !isSamlLoginFinishURL {
		return redirectURI
	}

	q := u.Query()
	q.Set(queryAuthenticationCode, e.ID)
	u.RawQuery = q.Encode()
	return u.String()
}

func (r *UIService) GetAuthenticationInfoID(req *http.Request) (string, bool) {
	code := req.FormValue(queryAuthenticationCode)
	if code != "" {
		return code, true
	}
	return "", false
}
