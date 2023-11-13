package webapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/tester"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/setutil"
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
)

//go:generate mockgen -source=authflow_controller.go -destination=authflow_controller_mock_test.go -package webapp

type AuthflowControllerHandler func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error
type AuthflowControllerErrorHandler func(w http.ResponseWriter, r *http.Request, err error) error

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

func GetXStepFromQuery(r *http.Request) string {
	xStep := r.URL.Query().Get(webapp.AuthflowQueryKey)
	return xStep
}

type AuthflowOAuthCallbackResponse struct {
	State            string
	Code             string
	Error            string
	ErrorDescription string
	ErrorURI         string
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

func (c *AuthflowController) HandleStartOfFlow(
	w http.ResponseWriter,
	r *http.Request,
	opts webapp.SessionOptions,
	flowReference authflow.FlowReference,
	handlers *AuthflowControllerHandlers,
	input interface{}) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s, err := c.getOrCreateWebSession(w, r, opts)
	if err != nil {
		c.Logger.WithError(err).Errorf("failed to get or create web session")
		c.renderError(w, r, err)
		return
	}

	handleWithScreen := func(screen *webapp.AuthflowScreenWithFlowResponse) {
		handler := c.makeHTTPHandler(s, screen, handlers)
		handler.ServeHTTP(w, r)
	}

	screen, err := c.GetScreen(s, GetXStepFromQuery(r))
	if err != nil {
		if errors.Is(err, authflow.ErrFlowNotFound) {
			screen, err := c.createScreen(r, s, flowReference, input)
			if err != nil {
				c.Logger.WithError(err).Errorf("failed to create screen")
				c.renderError(w, r, err)
				return
			}

			// We used to redirect to the same page with x_step added.
			// But that could be blocked by mobile Safari 17 private mode.
			// See https://github.com/authgear/authgear-server/issues/3470
			handleWithScreen(screen)
			return
		}

		c.Logger.WithError(err).Errorf("failed to get screen")
		c.renderError(w, r, err)
		return
	}

	err = c.checkPath(w, r, s, screen)
	if err != nil {
		c.renderError(w, r, err)
		return
	}

	handleWithScreen(screen)
}

func (c *AuthflowController) HandleOAuthCallback(w http.ResponseWriter, r *http.Request, callbackResponse AuthflowOAuthCallbackResponse) {
	state, err := webapp.DecodeAuthflowOAuthState(callbackResponse.State)
	if err != nil {
		c.Logger.WithError(err).Errorf("failed to get screen")
		c.renderError(w, r, err)
		return
	}

	s, err := c.getWebSession(r)
	if err != nil {
		if !apierrors.IsKind(err, webapp.WebUIInvalidSession) {
			c.Logger.WithError(err).Errorf("failed to get web session")
		}
		c.renderError(w, r, err)
		return
	}

	screen, err := c.GetScreen(s, state.XStep)
	if err != nil {
		c.Logger.WithError(err).Errorf("failed to get screen")
		c.renderError(w, r, err)
		return
	}

	input := map[string]interface{}{}
	switch {
	case callbackResponse.Code != "":
		input["code"] = callbackResponse.Code
	case callbackResponse.Error != "":
		input["error"] = callbackResponse.Error
		input["error_description"] = callbackResponse.ErrorDescription
		input["error_uri"] = callbackResponse.ErrorURI
	}
	result, err := c.AdvanceWithInput(r, s, screen, input)
	if err != nil {
		u, parseURLErr := url.Parse(state.ErrorRedirectURI)
		if parseURLErr != nil {
			panic(parseURLErr)
		}

		c.makeErrorResult(w, r, *u, err).WriteResponse(w, r)
		return
	}

	result.WriteResponse(w, r)
	return
}

func (c *AuthflowController) HandleResumeOfFlow(
	w http.ResponseWriter,
	r *http.Request,
	opts webapp.SessionOptions,
	handlers *AuthflowControllerHandlers,
	input map[string]interface{},
	errorHandler *AuthflowControllerErrorHandler,
) {

	handleError := func(err error) {
		if errorHandler != nil {
			err = (*errorHandler)(w, r, err)
			if err != nil {
				c.renderError(w, r, err)
			}
		}
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s, err := c.getOrCreateWebSession(w, r, opts)
	if err != nil {
		c.Logger.WithError(err).Errorf("failed to get or create web session")
		handleError(err)
		return
	}

	output, err := c.feedInput("", input)
	if err != nil {
		c.Logger.WithError(err).Errorf("failed to resume flow")
		handleError(err)
		return
	}

	screen, err := c.createScreenWithOutput(r, s, output, "")
	if err != nil {
		c.Logger.WithError(err).Errorf("failed to create screen")
		handleError(err)
		return
	}

	err = c.checkPath(w, r, s, screen)
	if err != nil {
		handleError(err)
		return
	}

	handler := c.makeHTTPHandler(s, screen, handlers)
	handler.ServeHTTP(w, r)
}

func (c *AuthflowController) HandleStep(w http.ResponseWriter, r *http.Request, handlers *AuthflowControllerHandlers) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s, err := c.getWebSession(r)
	if err != nil {
		if !apierrors.IsKind(err, webapp.WebUIInvalidSession) {
			c.Logger.WithError(err).Errorf("failed to get web session")
		}
		c.renderError(w, r, err)
		return
	}

	screen, err := c.GetScreen(s, GetXStepFromQuery(r))
	if err != nil {
		c.Logger.WithError(err).Errorf("failed to get screen")
		c.renderError(w, r, err)
		return
	}

	err = c.checkPath(w, r, s, screen)
	if err != nil {
		c.renderError(w, r, err)
		return
	}

	handler := c.makeHTTPHandler(s, screen, handlers)
	handler.ServeHTTP(w, r)
}

func (c *AuthflowController) HandleWithoutFlow(w http.ResponseWriter, r *http.Request, handlers *AuthflowControllerHandlers) {
	var session *webapp.Session
	s, err := c.getWebSession(r)
	if err != nil {
		if !apierrors.IsKind(err, webapp.WebUIInvalidSession) {
			c.Logger.WithError(err).Errorf("failed to get web session")
		}
	} else {
		session = s
	}

	handler := c.makeHTTPHandler(session, nil, handlers)
	handler.ServeHTTP(w, r)
}

func (c *AuthflowController) getWebSession(r *http.Request) (*webapp.Session, error) {
	s := webapp.GetSession(r.Context())
	if s == nil {
		return nil, webapp.ErrSessionNotFound
	}
	return s, nil
}

func (c *AuthflowController) getOrCreateWebSession(w http.ResponseWriter, r *http.Request, opts webapp.SessionOptions) (*webapp.Session, error) {
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

func (c *AuthflowController) GetScreen(s *webapp.Session, xStep string) (*webapp.AuthflowScreenWithFlowResponse, error) {
	if s.Authflow == nil {
		return nil, authflow.ErrFlowNotFound
	}

	screen, ok := s.Authflow.AllScreens[xStep]
	if !ok {
		return nil, authflow.ErrFlowNotFound
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

func (c *AuthflowController) createAuthflow(r *http.Request, s *webapp.Session, flowReference authflow.FlowReference) (*authflow.ServiceOutput, error) {
	flow, err := authflow.InstantiateFlow(flowReference, jsonpointer.T{})
	if err != nil {
		return nil, err
	}

	var sessionOptionsFromOAuth *authflow.SessionOptions
	if s.OAuthSessionID != "" {
		sessionOptionsFromOAuth, err = c.makeSessionOptionsFromOAuth(s.OAuthSessionID)
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

// ReplaceScreen is for switching flow.
func (c *AuthflowController) ReplaceScreen(r *http.Request, s *webapp.Session, flowReference authflow.FlowReference, input map[string]interface{}) (result *webapp.Result, err error) {
	var screen *webapp.AuthflowScreenWithFlowResponse
	result = &webapp.Result{}

	output, err := c.createAuthflow(r, s, flowReference)
	if err != nil {
		return
	}

	flowResponse := output.ToFlowResponse()
	emptyXStep := ""
	var emptyInput map[string]interface{}
	screen = webapp.NewAuthflowScreenWithFlowResponse(&flowResponse, emptyXStep, emptyInput)
	s.RememberScreen(screen)

	output, screen, err = c.takeBranchRecursively(s, screen)
	if err != nil {
		return
	}

	now := c.Clock.NowUTC()
	s.UpdatedAt = now
	err = c.Sessions.Update(s)
	if err != nil {
		return
	}

	result, err = c.AdvanceWithInput(r, s, screen, input)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *AuthflowController) Restart(s *webapp.Session) (result *webapp.Result, err error) {
	s.Authflow = nil
	// To be safe, also clear any interaction.
	s.Steps = []webapp.SessionStep{}
	now := c.Clock.NowUTC()
	s.UpdatedAt = now
	err = c.Sessions.Update(s)
	if err != nil {
		return
	}

	u := webapp.MakeRelativeURL("/flows/select_account", url.Values{})
	result = &webapp.Result{
		RedirectURI: u.String(),
		RemoveQueries: setutil.Set[string]{
			"x_step": struct{}{},
		},
	}
	return
}

func (c *AuthflowController) createScreenWithOutput(
	r *http.Request,
	s *webapp.Session,
	output *authflow.ServiceOutput,
	prevXStep string,
) (*webapp.AuthflowScreenWithFlowResponse, error) {
	flowResponse := output.ToFlowResponse()
	var emptyInput map[string]interface{}
	screen := webapp.NewAuthflowScreenWithFlowResponse(&flowResponse, prevXStep, emptyInput)
	s.RememberScreen(screen)

	output, screen, err := c.takeBranchRecursively(s, screen)
	if err != nil {
		return nil, err
	}

	now := c.Clock.NowUTC()
	s.UpdatedAt = now
	err = c.Sessions.Update(s)
	if err != nil {
		return nil, err
	}

	return screen, nil
}

func (c *AuthflowController) createScreen(
	r *http.Request,
	s *webapp.Session,
	flowReference authflow.FlowReference,
	input interface{}) (screen *webapp.AuthflowScreenWithFlowResponse, err error) {
	output1, err := c.createAuthflow(r, s, flowReference)
	if err != nil {
		return
	}

	screen, err = c.createScreenWithOutput(r, s, output1, "")
	if err != nil {
		return
	}

	if input != nil {
		output2, err := c.feedInput(output1.Flow.StateToken, input)
		if err != nil {
			return nil, err
		}
		screen, err = c.createScreenWithOutput(r, s, output2, screen.Screen.StateToken.XStep)
		if err != nil {
			return nil, err
		}
	}

	return
}

// AdvanceWithInputs is for feeding multiple inputs that would advance the flow.
func (c *AuthflowController) AdvanceWithInputs(
	r *http.Request,
	s *webapp.Session,
	screen0 *webapp.AuthflowScreenWithFlowResponse,
	inputs []map[string]interface{},
) (*webapp.Result, error) {
	result := &webapp.Result{}

	prevXStep := screen0.Screen.StateToken.XStep
	currentScreen := screen0

	for _, input := range inputs {
		output1, err := c.feedInput(currentScreen.Screen.StateToken.StateToken, input)
		if err != nil {
			return nil, err
		}

		result.Cookies = append(result.Cookies, output1.Cookies...)
		flowResponse1 := output1.ToFlowResponse()

		screen1 := webapp.NewAuthflowScreenWithFlowResponse(&flowResponse1, prevXStep, input)
		s.RememberScreen(screen1)
		currentScreen = screen1

		output2, screen2, err := c.takeBranchRecursively(s, screen1)
		if err != nil {
			return nil, err
		}
		currentScreen = screen2

		if output2 != nil {
			result.Cookies = append(result.Cookies, output2.Cookies...)
		}
		prevXStep = screen2.Screen.StateToken.XStep

		flowResponse2 := *screen2.StateTokenFlowResponse

		if flowResponse2.Action.Type == authflow.FlowActionTypeFinished {
			result.RemoveQueries = setutil.Set[string]{
				"x_step": struct{}{},
			}
			result.NavigationAction = "redirect"
			result.RedirectURI = c.deriveFinishRedirectURI(r, s, &flowResponse2)

			switch flowResponse2.Type {
			case authflow.FlowTypeLogin:
				fallthrough
			case authflow.FlowTypePromote:
				fallthrough
			case authflow.FlowTypeSignup:
				fallthrough
			case authflow.FlowTypeSignupLogin:
				fallthrough
			case authflow.FlowTypeReauth:
				// Forget the session.
				err = c.Sessions.Delete(s.ID)
				if err != nil {
					return nil, err
				}
				result.Cookies = append(result.Cookies, c.Cookies.ClearCookie(c.SessionCookie.Def))
				// Reset visitor ID.
				result.Cookies = append(result.Cookies, c.Cookies.ClearCookie(webapp.VisitorIDCookieDef))
			default:
				// Do nothing for other flows
			}
			// Ignore remaining inputs
			return result, nil
		} else {
			now := c.Clock.NowUTC()
			s.UpdatedAt = now
			err = c.Sessions.Update(s)
			if err != nil {
				return nil, err
			}
		}
	}

	currentScreen.Navigate(r, s.ID, result)

	return result, nil
}

// AdvanceWithInput is same as AdvanceWithInputs but only allow one input.
func (c *AuthflowController) AdvanceWithInput(
	r *http.Request,
	s *webapp.Session,
	screen0 *webapp.AuthflowScreenWithFlowResponse,
	input map[string]interface{},
) (result *webapp.Result, err error) {
	return c.AdvanceWithInputs(r, s, screen0, []map[string]interface{}{input})
}

// UpdateWithInput is for feeding an input that would just update the current node.
// One application is resend.
func (c *AuthflowController) UpdateWithInput(r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse, input map[string]interface{}) (result *webapp.Result, err error) {
	result = &webapp.Result{}

	output, err := c.feedInput(screen.Screen.StateToken.StateToken, input)
	if err != nil {
		return
	}

	result.Cookies = append(result.Cookies, output.Cookies...)
	newF := output.ToFlowResponse()
	newScreen := webapp.UpdateAuthflowScreenWithFlowResponse(screen, &newF)
	s.RememberScreen(newScreen)

	now := c.Clock.NowUTC()
	s.UpdatedAt = now
	err = c.Sessions.Update(s)
	if err != nil {
		return
	}

	if output != nil {
		result.Cookies = append(result.Cookies, output.Cookies...)
	}

	newScreen.Navigate(r, s.ID, result)
	return
}

func (c *AuthflowController) takeBranchRecursively(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (output *authflow.ServiceOutput, newScreen *webapp.AuthflowScreenWithFlowResponse, err error) {
	for screen.HasBranchToTake() {
		// Take the first branch, and first channel by default.
		var zeroIndex int
		var zeroChannel model.AuthenticatorOOBChannel
		takeBranchResult := screen.TakeBranch(zeroIndex, zeroChannel)

		switch takeBranchResult := takeBranchResult.(type) {
		// This taken branch does not require an input to select.
		case webapp.TakeBranchResultSimple:
			s.RememberScreen(takeBranchResult.Screen)
			screen = takeBranchResult.Screen
		// This taken branch require an input to select.
		case webapp.TakeBranchResultInput:
			output, err = c.feedInput(screen.Screen.StateToken.StateToken, takeBranchResult.Input)
			if err != nil {
				return
			}

			flowResponse := output.ToFlowResponse()
			screen = takeBranchResult.NewAuthflowScreenFull(&flowResponse)
			s.RememberScreen(screen)
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
	if err != nil && !errors.Is(err, authflow.ErrEOF) {
		return nil, err
	}

	return output, nil
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

	// 1. Use predefined redirect path of the flow
	// 2. Use the finish_redirect_uri from authflow. (To return to /oauth2/consent)
	// 3. Use redirect URI in webapp.Session.
	// 4. DerivePostLoginRedirectURIFromRequest

	path := webapp.DeriveAuthflowFinishPath(f)
	if path != "" {
		return path
	}

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

		IDToken:                  uiInfo.IDTokenHint,
		SuppressIDPSessionCookie: uiInfo.SuppressIDPSessionCookie,
		UserIDHint:               uiInfo.UserIDHint,
		LoginHint:                uiInfo.LoginHint,
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

func (c *AuthflowController) makeHTTPHandler(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse, handlers *AuthflowControllerHandlers) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error

		switch r.Method {
		case http.MethodGet:
			err = handlers.GetHandler(s, screen)
		case http.MethodPost:
			xAction := r.FormValue("x_action")
			switch xAction {
			case "take_branch":
				err = c.takeBranch(w, r, s, screen)
			default:
				handler, ok := handlers.PostHandlers[xAction]
				if !ok {
					http.Error(w, "Unknown action", http.StatusBadRequest)
					return
				}

				err = handler(s, screen)
			}
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if err != nil {
			if apierrors.IsAPIError(err) {
				c.renderError(w, r, err)
			} else {
				panic(err)
			}
		}
	})
}

func (c *AuthflowController) takeBranch(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
	xStepAtBranch := screen.Screen.BranchStateToken.XStep
	screen, err := c.GetScreen(s, xStepAtBranch)
	if err != nil {
		return err
	}

	indexStr := r.Form.Get("x_index")
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return err
	}
	channel := r.Form.Get("x_channel")

	takeBranchResult := screen.TakeBranch(index, model.AuthenticatorOOBChannel(channel))

	var output *authflow.ServiceOutput
	var newScreen *webapp.AuthflowScreenWithFlowResponse
	switch takeBranchResult := takeBranchResult.(type) {
	// This taken branch does not require an input to select.
	case webapp.TakeBranchResultSimple:
		s.RememberScreen(takeBranchResult.Screen)
		newScreen = takeBranchResult.Screen
	// This taken branch require an input to select.
	case webapp.TakeBranchResultInput:
		output, err = c.feedInput(screen.Screen.StateToken.StateToken, takeBranchResult.Input)
		if err != nil {
			return err
		}

		flowResponse := output.ToFlowResponse()
		newScreen = takeBranchResult.NewAuthflowScreenFull(&flowResponse)
		s.RememberScreen(newScreen)
	}

	now := c.Clock.NowUTC()
	s.UpdatedAt = now
	err = c.Sessions.Update(s)
	if err != nil {
		return err
	}

	result := &webapp.Result{}
	if output != nil {
		result.Cookies = append(result.Cookies, output.Cookies...)
	}

	newScreen.Navigate(r, s.ID, result)
	result.WriteResponse(w, r)
	return nil
}

func (c *AuthflowController) makeErrorResult(w http.ResponseWriter, r *http.Request, u url.URL, err error) *webapp.Result {
	apierror := apierrors.AsAPIError(err)

	recoverable := func() *webapp.Result {
		cookie, err := c.ErrorCookie.SetRecoverableError(r, apierror)
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
		err := c.ErrorCookie.SetNonRecoverableError(result, apierror)
		if err != nil {
			panic(err)
		}

		return result
	}

	switch {
	case errors.Is(err, authflow.ErrFlowNotFound):
		u.Path = "/errors/error"
		return nonRecoverable()
	case user.IsAccountStatusError(err):
		u.Path = webapp.AuthflowRouteAccountStatus
		return nonRecoverable()
	case errors.Is(err, api.ErrNoAuthenticator):
		u.Path = webapp.AuthflowRouteNoAuthenticator
		return nonRecoverable()
	case apierrors.IsKind(err, webapp.WebUIInvalidSession):
		// Show WebUIInvalidSession error in different page.
		u.Path = "/errors/error"
		return nonRecoverable()
	case r.Method == http.MethodGet:
		// If the request method is Get, avoid redirect back to the same path
		// which causes infinite redirect loop
		u.Path = "/errors/error"
		return nonRecoverable()
	default:
		return recoverable()
	}
}

func (c *AuthflowController) renderError(w http.ResponseWriter, r *http.Request, err error) {
	c.makeErrorResult(w, r, *r.URL, err).WriteResponse(w, r)
}

func (c *AuthflowController) checkPath(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
	// We derive the intended path of the screen,
	// and check if the paths match.
	result := &webapp.Result{}
	screen.Navigate(r, s.ID, result)
	redirectURI := result.RedirectURI

	if redirectURI == "" {
		panic(fmt.Errorf("expected Navigate to set RedirectURI"))
	}

	u, err := url.Parse(redirectURI)
	if err != nil {
		return err
	}

	if u.Path != r.URL.Path {
		// We do not know what causes the mismatch.
		// Maybe x_step was tempered.
		c.Logger.Warningln("path mismatch")
		return webapp.ErrInvalidSession
	}

	return nil
}
