package webapp

import (
	"context"
	"errors"
	"net/http"

	"net/url"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
)

type ErrorRendererAuthflowV2Navigator interface {
	NavigateNonRecoverableError(r *http.Request, u *url.URL, e error)
}

type ErrorRenderer struct {
	ErrorService        *webapp.ErrorService
	AuthflowV2Navigator ErrorRendererAuthflowV2Navigator
}

func (s *ErrorRenderer) RenderError(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	s.renderAuthflowError(ctx, w, r, err)
}

func (s *ErrorRenderer) MakeAuthflowErrorResult(ctx context.Context, w http.ResponseWriter, r *http.Request, u url.URL, err error) *webapp.Result {
	apierror := apierrors.AsAPIError(err)

	recoverable := func() *webapp.Result {
		cookie, err := s.ErrorService.SetRecoverableError(ctx, r, apierror)
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
		s.AuthflowV2Navigator.NavigateNonRecoverableError(r, &u, err)
		return nonRecoverable()
	default:
		return recoverable()
	}
}

func (s *ErrorRenderer) renderAuthflowError(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	s.MakeAuthflowErrorResult(ctx, w, r, *r.URL, err).WriteResponse(w, r)
}
