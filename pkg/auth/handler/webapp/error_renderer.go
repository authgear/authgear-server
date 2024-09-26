package webapp

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type ErrorRendererUIImplementationService interface {
	GetUIImplementation() config.UIImplementation
}

type ErrorRendererAuthflowV1Navigator interface {
	NavigateNonRecoverableError(r *http.Request, u *url.URL, e error)
}
type ErrorRendererAuthflowV2Navigator interface {
	NavigateNonRecoverableError(r *http.Request, u *url.URL, e error)
}

type ErrorRenderer struct {
	ErrorService            *webapp.ErrorService
	UIImplementationService ErrorRendererUIImplementationService

	AuthflowV1Navigator ErrorRendererAuthflowV1Navigator
	AuthflowV2Navigator ErrorRendererAuthflowV2Navigator
}

func (s *ErrorRenderer) RenderError(w http.ResponseWriter, r *http.Request, err error) {
	uiImpl := s.UIImplementationService.GetUIImplementation()
	switch uiImpl {
	case config.UIImplementationInteraction:
		s.renderInteractionError(w, r, err)
	case config.UIImplementationAuthflow:
		fallthrough
	case config.UIImplementationAuthflowV2:
		s.renderAuthflowError(w, r, err)
	default:
		panic(fmt.Errorf("unknown ui implementation %s", uiImpl))
	}
}

func (s *ErrorRenderer) renderInteractionError(w http.ResponseWriter, r *http.Request, err error) {
	apierror := apierrors.AsAPIError(err)

	// Show WebUIInvalidSession error in different page.
	u := *r.URL
	// If the request method is Get, avoid redirect back to the same path
	// which causes infinite redirect loop
	if r.Method == http.MethodGet {
		u.Path = "/errors/error"
	}
	if apierror.Reason == webapp.WebUIInvalidSession.Reason {
		u.Path = "/errors/error"
	}

	cookie, err := s.ErrorService.SetRecoverableError(r, apierror)
	if err != nil {
		panic(err)
	}
	result := webapp.Result{
		RedirectURI:      u.String(),
		NavigationAction: "replace",
		Cookies:          []*http.Cookie{cookie},
	}
	result.WriteResponse(w, r)
}

func (s *ErrorRenderer) MakeAuthflowErrorResult(w http.ResponseWriter, r *http.Request, u url.URL, err error) *webapp.Result {
	apierror := apierrors.AsAPIError(err)

	recoverable := func() *webapp.Result {
		cookie, err := s.ErrorService.SetRecoverableError(r, apierror)
		if err != nil {
			panic(err)
		}

		result := &webapp.Result{
			RedirectURI:      u.String(),
			NavigationAction: "replace",
			Cookies:          []*http.Cookie{cookie},
		}

		return result
	}

	nonRecoverable := func() *webapp.Result {
		result := &webapp.Result{
			RedirectURI:      u.String(),
			NavigationAction: "replace",
		}
		err := s.ErrorService.SetNonRecoverableError(result, apierror)
		if err != nil {
			panic(err)
		}

		return result
	}

	switch {
	case apierror.Reason == "AuthenticationFlowNoPublicSignup":
		fallthrough
	case errors.Is(err, authflow.ErrFlowNotFound):
		fallthrough
	case user.IsAccountStatusError(err):
		fallthrough
	case errors.Is(err, api.ErrNoAuthenticator):
		fallthrough
	case apierrors.IsKind(err, webapp.WebUIInvalidSession):
		fallthrough
	case r.Method == http.MethodGet && u.Path == r.URL.Path:
		// Infinite loop might occur if it is a GET request with the same route
		switch s.UIImplementationService.GetUIImplementation() {
		case config.UIImplementationAuthflow:
			s.AuthflowV1Navigator.NavigateNonRecoverableError(r, &u, err)
		case config.UIImplementationAuthflowV2:
			s.AuthflowV2Navigator.NavigateNonRecoverableError(r, &u, err)
		}
		return nonRecoverable()
	default:
		return recoverable()
	}
}

func (s *ErrorRenderer) renderAuthflowError(w http.ResponseWriter, r *http.Request, err error) {
	s.MakeAuthflowErrorResult(w, r, *r.URL, err).WriteResponse(w, r)
}
