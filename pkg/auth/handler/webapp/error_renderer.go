package webapp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"net/url"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

type ErrorRendererUIImplementationService interface {
	GetUIImplementation() config.UIImplementation
}

type ErrorRendererAuthflowV2Navigator interface {
	NavigateNonRecoverableError(r *http.Request, u *url.URL, e error)
}

var ErrorRendererLogger = slogutil.NewLogger("error_renderer")

type ErrorRenderer struct {
	ErrorService            *webapp.ErrorService
	UIImplementationService ErrorRendererUIImplementationService
	Renderer                Renderer
	BaseViewModel           *viewmodels.BaseViewModeler

	AuthflowV2Navigator ErrorRendererAuthflowV2Navigator
}

func (s *ErrorRenderer) RenderError(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	uiImpl := s.UIImplementationService.GetUIImplementation()
	switch uiImpl {
	case config.UIImplementationAuthflowV2:
		s.renderAuthflowError(ctx, w, r, err)
	default:
		panic(fmt.Errorf("unknown ui implementation %s", uiImpl))
	}
}

// TODO(tung)
func (s *ErrorRenderer) renderInteractionError(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	apierror := apierrors.AsAPIErrorWithContext(ctx, err)

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

func (h *ErrorRenderer) GetErrorData(ctx context.Context, r *http.Request, w http.ResponseWriter, err error) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, w)
	baseViewModel.SetError(ctx, err)
	viewmodels.Embed(data, baseViewModel)
	return data, nil
}

func (s *ErrorRenderer) makeSessionCompletedErrorResult(ctx context.Context, w http.ResponseWriter, r *http.Request, u url.URL, apierr *apierrors.APIError) httputil.Result {
	if r.Method == http.MethodGet {
		data, err := s.GetErrorData(ctx, r, w, apierr)
		if err != nil {
			return s.makeNonRecoverableResult(ctx, u, apierr)
		}
		return &HTMLResult{
			Template: TemplateV2WebFatalErrorHTML,
			Data:     data,
			Renderer: s.Renderer,
		}
	} else {
		// If it is not a POST request, redirect once to the same url with GET, and display error there.
		result := &webapp.Result{
			RedirectURI:      u.String(),
			NavigationAction: webapp.NavigationActionReplace,
		}
		return result
	}
}

func (s *ErrorRenderer) makeNonRecoverableResult(ctx context.Context, u url.URL, apierr *apierrors.APIError) httputil.Result {
	result := &webapp.Result{
		RedirectURI:      u.String(),
		NavigationAction: webapp.NavigationActionReplace,
	}
	err := s.ErrorService.SetNonRecoverableError(ctx, result, apierr)
	if err != nil {
		panic(err)
	}

	return result
}

func (s *ErrorRenderer) MakeAuthflowErrorResult(ctx context.Context, w http.ResponseWriter, r *http.Request, u url.URL, err error) httputil.Result {
	logger := ErrorRendererLogger.GetLogger(ctx)
	apierr := apierrors.AsAPIErrorWithContext(ctx, err)

	if apierrors.IsKind(apierr, apierrors.UnexpectedError) {
		logger.WithError(err).Error(ctx, "unexpected error")
	}

	recoverable := func() *webapp.Result {
		cookie, err := s.ErrorService.SetRecoverableError(ctx, r, apierr)
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

	switch {
	case apierrors.IsKind(err, webapp.WebUISessionCompleted):
		return s.makeSessionCompletedErrorResult(ctx, w, r, u, apierr)
	case apierrors.IsKind(err, api.OAuthProviderMissingCredentials):
		fallthrough
	case apierr.Reason == "AuthenticationFlowNoPublicSignup":
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
		case config.UIImplementationAuthflowV2:
			s.AuthflowV2Navigator.NavigateNonRecoverableError(r, &u, err)
		}
		apierr = apierr.CloneWithAdditionalInfo(apierrors.Details{
			"FromURL": r.URL.String(),
		})
		return s.makeNonRecoverableResult(ctx, u, apierr)
	default:
		return recoverable()
	}
}

func (s *ErrorRenderer) renderAuthflowError(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	s.MakeAuthflowErrorResult(ctx, w, r, *r.URL, err).WriteResponse(w, r)
}

type HTMLResult struct {
	Renderer Renderer
	Template *template.HTML
	Data     interface{}
}

func (re *HTMLResult) WriteResponse(rw http.ResponseWriter, r *http.Request) {
	re.Renderer.RenderHTML(rw, r, TemplateV2WebFatalErrorHTML, re.Data)
}

func (re *HTMLResult) IsInternalError() bool {
	return false
}
