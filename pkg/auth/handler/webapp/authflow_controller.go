package webapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/tester"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type AuthflowControllerHandler func() error

type AuthflowControllerHandlers struct {
	GetHandler   AuthflowControllerHandler
	PostHandlers map[string]AuthflowControllerHandler
}

func (h *AuthflowControllerHandlers) Get(f AuthflowControllerHandler) {
	h.GetHandler = f
}

func (h *AuthflowControllerHandlers) PostAction(action string, f AuthflowControllerHandler) {
	if h.PostHandlers == nil {
		h.PostHandlers = make(map[string]AuthflowControllerHandler)
	}
	h.PostHandlers[action] = f
}

type AuthflowControllerSessionStore interface {
	Get(id string) (*webapp.Session, error)
	Create(session *webapp.Session) (err error)
	Update(session *webapp.Session) (err error)
	Delete(id string) (err error)
}

type AuthflowControllerAuthflowService interface {
	CreateNewFlow(intent authflow.PublicFlow, sessionOptions *authflow.SessionOptions) (*authflow.ServiceOutput, error)
	Get(stateToken string) (*authflow.ServiceOutput, error)
	FeedInput(stateToken string, rawMessage json.RawMessage) (*authflow.ServiceOutput, error)
}

type AuthflowControllerOAuthSessionService interface {
	Get(entryID string) (*oauthsession.Entry, error)
}

type AuthflowControllerUIInfoResolver interface {
	ResolveForUI(r protocol.AuthorizationRequest) (*oidc.UIInfo, error)
}

type AuthflowController struct {
	TesterEndpointsProvider tester.EndpointsProvider
	ErrorCookie             *webapp.ErrorCookie
	TrustProxy              config.TrustProxy
	Clock                   clock.Clock

	Cookies       CookieManager
	Sessions      AuthflowControllerSessionStore
	SessionCookie webapp.SessionCookieDef

	Authflows AuthflowControllerAuthflowService

	OAuthSessions  AuthflowControllerOAuthSessionService
	UIInfoResolver AuthflowControllerUIInfoResolver
}

func (c *AuthflowController) GetOrCreateWebSession(w http.ResponseWriter, r *http.Request, opts webapp.SessionOptions) (*webapp.Session, error) {
	now := c.Clock.NowUTC()
	s := webapp.GetSession(r.Context())
	if s != nil {
		return s, nil
	}

	o := opts
	o.UpdatedAt = now

	s = webapp.NewSession(o)
	err := c.Sessions.Create(s)
	if err != nil {
		return nil, err
	}

	cookie := c.Cookies.ValueCookie(c.SessionCookie.Def, s.ID)
	httputil.UpdateCookie(w, cookie)

	return s, nil
}

func (c *AuthflowController) GetAuthflow(r *http.Request, s *webapp.Session) (*authflow.FlowResponse, error) {
	if s.Authflow == nil {
		return nil, authflow.ErrFlowNotFound
	}

	xStep := r.URL.Query().Get("x_step")

	var state *webapp.AuthflowState

	// entrypoint does not have x_step.
	if xStep == "" {
		state = s.Authflow.InitialState
	} else {
		var ok bool
		state, ok = s.Authflow.AllStates[xStep]
		if !ok {
			return nil, authflow.ErrFlowNotFound
		}
	}

	output, err := c.Authflows.Get(state.StateToken)
	if err != nil {
		return nil, err
	}

	flowResponse := output.ToFlowResponse()
	return &flowResponse, nil
}

func (c *AuthflowController) CreateAuthflow(r *http.Request, oauthSessionID string, flowReference authflow.FlowReference) (*authflow.FlowResponse, error) {
	flow, err := authflow.InstantiateFlow(flowReference)
	if err != nil {
		return nil, err
	}

	var sessionOptionsFromOAuth *authflow.SessionOptions
	if oauthSessionID != "" {
		sessionOptionsFromOAuth, err = c.makeSessionOptionsFromOAuth(oauthSessionID)
		if errors.Is(err, oauthsession.ErrNotFound) {
			// Ignore this error.
		} else if err != nil {
			return nil, err
		}
	}

	sessionOptionsFromQuery := c.makeSessionOptionsFromQuery(r)

	// The query overrides the cookie.
	sessionOptions := sessionOptionsFromOAuth.PartiallyMergeFrom(sessionOptionsFromQuery)

	output, err := c.Authflows.CreateNewFlow(flow, sessionOptions)
	if err != nil {
		return nil, err
	}

	flowResponse := output.ToFlowResponse()
	return &flowResponse, nil
}

func (c *AuthflowController) GetOrCreateAuthflow(r *http.Request, s *webapp.Session, flowReference authflow.FlowReference, checkFn func(*authflow.FlowResponse) bool) (*authflow.FlowResponse, error) {
	flowResponse, err := c.GetAuthflow(r, s)
	if err != nil {
		// Return critical error.
		if !errors.Is(err, authflow.ErrFlowNotFound) {
			return nil, err
		}
	}

	if flowResponse != nil {
		ok := checkFn(flowResponse)
		if ok {
			return flowResponse, nil
		}
	}

	// If we reach here, either flowResponse is nil or it does not the check.
	// We create a new authflow instead.
	flowResponse, err = c.CreateAuthflow(r, s.OAuthSessionID, flowReference)
	if err != nil {
		return nil, err
	}

	ok := checkFn(flowResponse)
	if !ok {
		panic(fmt.Errorf("a freshly created authflow must pass the check"))
	}

	af := webapp.NewAuthflow(flowResponse)
	s.Authflow = af
	now := c.Clock.NowUTC()
	s.UpdatedAt = now
	err = c.Sessions.Update(s)
	if err != nil {
		return nil, err
	}

	return flowResponse, nil
}

func (c *AuthflowController) makeSessionOptionsFromQuery(r *http.Request) *authflow.SessionOptions {
	q := r.URL.Query()
	return &authflow.SessionOptions{
		UILocales: q.Get("ui_locales"),
	}
}

func (c *AuthflowController) makeSessionOptionsFromOAuth(oauthSessionID string) (*authflow.SessionOptions, error) {
	entry, err := c.OAuthSessions.Get(oauthSessionID)
	if err != nil {
		return nil, err
	}
	req := entry.T.AuthorizationRequest

	uiInfo, err := c.UIInfoResolver.ResolveForUI(req)
	if err != nil {
		return nil, err
	}

	sessionOptions := &authflow.SessionOptions{
		OAuthSessionID: oauthSessionID,

		ClientID:    uiInfo.ClientID,
		RedirectURI: uiInfo.RedirectURI,
		Prompt:      uiInfo.Prompt,
		State:       uiInfo.State,
		XState:      uiInfo.XState,
		UILocales:   req.UILocalesRaw(),

		SuppressIDPSessionCookie: uiInfo.SuppressIDPSessionCookie,
		UserIDHint:               uiInfo.UserIDHint,
	}

	return sessionOptions, nil
}

func (c *AuthflowController) RedirectURI(r *http.Request) string {
	ruri := webapp.GetRedirectURI(r, bool(c.TrustProxy), "")
	// Consent screen will be skipped if redirect uri here is not empty
	// To workaround it we exclude tester endpoint here
	if ruri == c.TesterEndpointsProvider.TesterURL().String() {
		return ""
	}
	return ruri
}

func (c *AuthflowController) MakeHTTPHandler(handlers *AuthflowControllerHandlers) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error

		switch r.Method {
		case http.MethodGet:
			err = handlers.GetHandler()
		case http.MethodPost:
			xAction := r.FormValue("x_action")
			handler, ok := handlers.PostHandlers[xAction]
			if !ok {
				http.Error(w, "Unknown action", http.StatusBadRequest)
				return
			}

			err = handler()
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if apierrors.IsAPIError(err) {
			c.renderError(w, r, err)
		} else {
			panic(err)
		}
	})
}

func (c *AuthflowController) renderError(w http.ResponseWriter, r *http.Request, err error) {
	apierror := apierrors.AsAPIError(err)

	u := *r.URL

	// If the request method is Get, avoid redirect back to the same path
	// which causes infinite redirect loop
	if r.Method == http.MethodGet {
		u.Path = "/errors/error"
	}
	// Show WebUIInvalidSession error in different page.
	if apierror.Reason == webapp.WebUIInvalidSession.Reason {
		u.Path = "/errors/error"
	}

	cookie, err := c.ErrorCookie.SetError(r, apierror)
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
