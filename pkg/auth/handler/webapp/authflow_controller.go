package webapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/tester"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/setutil"
)

//go:generate mockgen -source=authflow_controller.go -destination=authflow_controller_mock_test.go -package webapp

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

type AuthflowControllerCookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
	ClearCookie(def *httputil.CookieDef) *http.Cookie
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

type AuthflowControllerOAuthClientResolver interface {
	ResolveClient(clientID string) *config.OAuthClientConfig
}

type AuthflowControllerLogger struct{ *log.Logger }

func NewAuthflowControllerLogger(lf *log.Factory) AuthflowControllerLogger {
	return AuthflowControllerLogger{lf.New("authflow_controller")}
}

type AuthflowController struct {
	Logger                  AuthflowControllerLogger
	TesterEndpointsProvider tester.EndpointsProvider
	ErrorCookie             *webapp.ErrorCookie
	TrustProxy              config.TrustProxy
	Clock                   clock.Clock

	Cookies       AuthflowControllerCookieManager
	Sessions      AuthflowControllerSessionStore
	SessionCookie webapp.SessionCookieDef

	Authflows AuthflowControllerAuthflowService

	OAuthSessions  AuthflowControllerOAuthSessionService
	UIInfoResolver AuthflowControllerUIInfoResolver

	UIConfig            *config.UIConfig
	OAuthClientResolver AuthflowControllerOAuthClientResolver
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

func (c *AuthflowController) GetScreen(r *http.Request, s *webapp.Session) (*webapp.AuthflowScreenWithFlowResponse, error) {
	if s.Authflow == nil {
		return nil, authflow.ErrFlowNotFound
	}

	xStep := r.URL.Query().Get(webapp.AuthflowQueryKey)

	var screen *webapp.AuthflowScreen

	// entrypoint does not have x_step.
	if xStep == "" {
		screen = s.Authflow.InitialScreen
	} else {
		var ok bool
		screen, ok = s.Authflow.AllScreens[xStep]
		if !ok {
			return nil, authflow.ErrFlowNotFound
		}
	}

	screenWithResponse := &webapp.AuthflowScreenWithFlowResponse{
		Screen: screen,
	}

	output, err := c.Authflows.Get(screen.StateToken.StateToken)
	if err != nil {
		return nil, err
	}
	flowResponse := output.ToFlowResponse()
	screenWithResponse.StateTokenFlowResponse = &flowResponse

	if screen.BranchStateToken != nil {
		output, err = c.Authflows.Get(screen.BranchStateToken.StateToken)
		if err != nil {
			return nil, err
		}

		flowResponse := output.ToFlowResponse()
		screenWithResponse.BranchStateTokenFlowResponse = &flowResponse
	}

	return screenWithResponse, nil
}

func (c *AuthflowController) createAuthflow(r *http.Request, oauthSessionID string, flowReference authflow.FlowReference) (*authflow.ServiceOutput, error) {
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

	return output, err
}

func (c *AuthflowController) GetOrCreateScreen(r *http.Request, s *webapp.Session, flowReference authflow.FlowReference, checkFn func(*authflow.FlowResponse) bool) (*webapp.AuthflowScreenWithFlowResponse, error) {
	screen, err := c.GetScreen(r, s)
	if err != nil {
		// Return critical error.
		if !errors.Is(err, authflow.ErrFlowNotFound) {
			return nil, err
		}
	}

	if screen != nil {
		ok := checkFn(screen.StateTokenFlowResponse)
		if ok {
			return screen, nil
		}
	}

	return c.createScreen(r, s, flowReference, checkFn)
}

func (c *AuthflowController) createScreen(r *http.Request, s *webapp.Session, flowReference authflow.FlowReference, checkFn func(*authflow.FlowResponse) bool) (*webapp.AuthflowScreenWithFlowResponse, error) {
	// If we reach here, either screen is nil or its flowResponse does not pass the check.
	// We create a new authflow instead.
	output, err := c.createAuthflow(r, s.OAuthSessionID, flowReference)
	if err != nil {
		return nil, err
	}

	flowResponse := output.ToFlowResponse()
	ok := checkFn(&flowResponse)
	if !ok {
		panic(fmt.Errorf("a freshly created authflow must pass the check"))
	}

	screen := webapp.NewAuthflowScreenWithFlowResponse(&flowResponse)
	af := webapp.NewAuthflow(flowResponse.ID, screen)
	s.Authflow = af

	output, newScreen, err := c.takeBranchIfNeeded(s, screen)
	if err != nil {
		return nil, err
	}

	now := c.Clock.NowUTC()
	s.UpdatedAt = now
	err = c.Sessions.Update(s)
	if err != nil {
		return nil, err
	}

	return newScreen, nil
}

func (c *AuthflowController) FeedInput(r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse, input interface{}) (result *webapp.Result, err error) {
	result = &webapp.Result{}

	defer func() {
		if err != nil {
			if !apierrors.IsAPIError(err) {
				c.Logger.WithField("path", r.URL.Path).WithError(err).Errorf("feed input: %v", err)
			}
			var errCookie *http.Cookie
			errCookie, err = c.ErrorCookie.SetError(r, apierrors.AsAPIError(err))
			if err != nil {
				result = nil
				return
			}
			result.IsInteractionErr = true
			result.Cookies = append(result.Cookies, errCookie)
			result.RedirectURI = c.deriveErrorRedirectURI(r, screen)
		}
	}()

	output, err := c.feedInput(screen.Screen.StateToken.StateToken, input)
	if err != nil {
		return
	}

	result.Cookies = append(result.Cookies, output.Cookies...)
	newF := output.ToFlowResponse()
	if newF.Action.Type == authflow.FlowActionTypeFinished {
		result.RemoveQueries = setutil.Set[string]{
			"x_step": struct{}{},
		}
		result.NavigationAction = "redirect"
		result.RedirectURI = c.deriveFinishRedirectURI(r, s, &newF)

		err = c.Sessions.Delete(s.ID)
		if err != nil {
			return
		}

		// Forget the session.
		result.Cookies = append(result.Cookies, c.Cookies.ClearCookie(c.SessionCookie.Def))
		// Reset visitor ID.
		result.Cookies = append(result.Cookies, c.Cookies.ClearCookie(webapp.VisitorIDCookieDef))
	} else {
		newScreen := webapp.NewAuthflowScreenWithFlowResponse(&newF)
		s.Authflow.RememberScreen(newScreen)

		output, newScreen, err = c.takeBranchIfNeeded(s, newScreen)
		if err != nil {
			return
		}

		now := c.Clock.NowUTC()
		s.UpdatedAt = now
		err = c.Sessions.Update(s)
		if err != nil {
			return
		}

		if output != nil {
			result.Cookies = append(result.Cookies, output.Cookies...)
		}

		newScreen.Navigate(r, result)
	}

	return
}

func (c *AuthflowController) takeBranchIfNeeded(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (output *authflow.ServiceOutput, newScreen *webapp.AuthflowScreenWithFlowResponse, err error) {
	if screen.HasBranchToTake() {
		// Take the first branch, and first channel by default.
		var zeroIndex int
		var zeroChannel model.AuthenticatorOOBChannel
		takeBranchResult := screen.TakeBranch(zeroIndex, zeroChannel)

		switch takeBranchResult := takeBranchResult.(type) {
		// This taken branch does not require an input to select.
		case webapp.TakeBranchResultSimple:
			s.Authflow.RememberScreen(takeBranchResult.Screen)
			newScreen = takeBranchResult.Screen
			return
		// This taken branch require an input to select.
		case webapp.TakeBranchResultInput:
			output, err = c.feedInput(screen.Screen.StateToken.StateToken, takeBranchResult.Input)
			if err != nil {
				return
			}

			flowResponse := output.ToFlowResponse()
			newScreen = takeBranchResult.NewAuthflowScreenFull(&flowResponse)
			s.Authflow.RememberScreen(newScreen)

			return
		}
	}

	newScreen = screen
	return
}

func (c *AuthflowController) feedInput(stateToken string, input interface{}) (*authflow.ServiceOutput, error) {
	rawMessageBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	rawMessage := json.RawMessage(rawMessageBytes)
	output, err := c.Authflows.FeedInput(stateToken, rawMessage)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (c *AuthflowController) deriveErrorRedirectURI(r *http.Request, screen *webapp.AuthflowScreenWithFlowResponse) string {
	// Just redirect to the current location.
	u := *r.URL
	u.Scheme = ""
	u.Host = ""
	return u.String()
}

func (c *AuthflowController) deriveFinishRedirectURI(r *http.Request, s *webapp.Session, f *authflow.FlowResponse) string {
	bytes, err := json.Marshal(f.Action.Data)
	if err != nil {
		panic(err)
	}

	var data map[string]interface{}
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		panic(err)
	}

	// 1. Use the finish_redirect_uri from authflow. (To return to /oauth2/consent)
	// 2. Use redirect URI in webapp.Session.
	// 3. DerivePostLoginRedirectURIFromRequest

	if finishRedirectURI, ok := data["finish_redirect_uri"].(string); ok {
		return finishRedirectURI
	}
	if s.RedirectURI != "" {
		return s.RedirectURI
	}

	postLoginRedirectURI := webapp.DerivePostLoginRedirectURIFromRequest(r, c.OAuthClientResolver, c.UIConfig)
	return webapp.GetRedirectURI(r, bool(c.TrustProxy), postLoginRedirectURI)
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
