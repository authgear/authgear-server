package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/dependency/sso"
	"github.com/authgear/authgear-server/pkg/core/crypto"
	"github.com/authgear/authgear-server/pkg/core/errors"
)

type SSOStateCodec interface {
	EncodeState(state sso.State) (string, error)
	DecodeState(encodedState string) (*sso.State, error)
}

type OAuthProviderFactory interface {
	NewOAuthProvider(alias string) sso.OAuthProvider
}

type OAuthService struct {
	StateProvider        StateProvider
	SSOStateCodec        SSOStateCodec
	OAuthProviderFactory OAuthProviderFactory
}

func (s *OAuthService) LoginOAuthProvider(w http.ResponseWriter, r *http.Request, providerAlias string) (writeResponse func(err error), err error) {
	var authURI string
	var state *State

	writeResponse = func(err error) {
		s.StateProvider.UpdateState(state, nil, err)
		if err != nil {
			RedirectToCurrentPath(w, r)
		} else {
			http.Redirect(w, r, authURI, http.StatusFound)
		}
	}

	oauthProvider := s.OAuthProviderFactory.NewOAuthProvider(providerAlias)
	if oauthProvider == nil {
		err = ErrOAuthProviderNotFound
		return
	}

	state = s.StateProvider.CreateState(r, nil, nil)

	// set hashed csrf cookies to sso state
	// callback will verify if the request has the same cookie
	cookie, err := r.Cookie(csrfCookieName)
	if err != nil || cookie.Value == "" {
		panic(errors.Newf("webapp: missing csrf cookies: %w", err))
	}
	hashedNonce := crypto.SHA256String(cookie.Value)
	webappSSOState := SSOState{}
	// Redirect back to the current page on error.
	q := r.URL.Query()
	q.Set("error_uri", r.URL.Path)
	webappSSOState.SetRequestQuery(q.Encode())
	ssoState := sso.State{
		Action:      "login",
		HashedNonce: hashedNonce,
		Extra:       webappSSOState,
	}
	encodedState, err := s.SSOStateCodec.EncodeState(ssoState)
	if err != nil {
		return
	}
	authURI, err = oauthProvider.GetAuthURL(ssoState, encodedState)
	return
}
