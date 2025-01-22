package webapp

import (
	"context"
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
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
)

type ErrorRendererUIImplementationService interface {
	GetUIImplementation() config.UIImplementation
}

type ErrorRendererAuthflowV2Navigator interface {
	NavigateNonRecoverableError(r *http.Request, u *url.URL, e error)
}

type ErrorRenderer struct {
	ErrorService            *webapp.ErrorService
	UIImplementationService ErrorRendererUIImplementationService

	AuthflowV2Navigator ErrorRendererAuthflowV2Navigator
}

func (s *ErrorRenderer) RenderError(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	uiImpl := s.UIImplementationService.GetUIImplementation()
	switch uiImpl {
	case config.UIImplementationInteraction:
		s.renderInteractionError(ctx, w, r, err)
	case config.UIImplementationAuthflowV2:
		s.renderAuthflowError(ctx, w, r, err)
	default:
		panic(fmt.Errorf("unknown ui implementation %s", uiImpl))
	}
}

func (s *ErrorRenderer) renderInteractionError(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
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

	cookie, err := s.ErrorService.SetRecoverableError(ctx, r, apierror)
	if err != nil {
		panic(err)
	}
	result := webapp.Result{
		RedirectURI:      u.String(),
		NavigationAction: webapp.NavigationActionReplace,
		Cookies:          []*http.Cookie{cookie},
	}
	result.WriteResponse(w, r)
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
			NavigationAction: webapp.NavigationActionReplace,
			Cookies:          []*http.Cookie{cookie},
		}

		return result
	}

	nonRecoverable := func() *webapp.Result {
		result := &webapp.Result{
			RedirectURI:      u.String(),
			NavigationAction: webapp.NavigationActionReplace,
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
	case apierrors.IsKind(err, smsapi.NoAvailableClient):
		fallthrough
	case r.Method == http.MethodGet && u.Path == r.URL.Path:
		// Infinite loop might occur if it is a GET request with the same route
		switch s.UIImplementationService.GetUIImplementation() {
		case config.UIImplementationAuthflowV2:
			s.AuthflowV2Navigator.NavigateNonRecoverableError(r, &u, err)
		}
		return nonRecoverable()
	default:
		return recoverable()
	}
}

func (s *ErrorRenderer) renderAuthflowError(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	s.MakeAuthflowErrorResult(ctx, w, r, *r.URL, err).WriteResponse(w, r)
}
