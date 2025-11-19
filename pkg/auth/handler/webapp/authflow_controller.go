package webapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"net/url"
	"strconv"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflowv2viewmodels "github.com/authgear/authgear-server/pkg/auth/handler/webapp/authflowv2/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlsession"
	"github.com/authgear/authgear-server/pkg/lib/tester"
	"github.com/authgear/authgear-server/pkg/lib/webappoauth"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/setutil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

//go:generate go tool mockgen -source=authflow_controller.go -destination=authflow_controller_mock_test.go -package webapp

type AuthflowControllerHandler func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error
type AuthflowControllerErrorHandler func(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) error
type AuthflowControllerInlinePreviewHandler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

type AuthflowControllerHandlers struct {
	GetHandler           AuthflowControllerHandler
	PostHandlers         map[string]AuthflowControllerHandler
	InlinePreviewHandler AuthflowControllerInlinePreviewHandler
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

func (h *AuthflowControllerHandlers) InlinePreview(f AuthflowControllerInlinePreviewHandler) {
	h.InlinePreviewHandler = f
}

type AuthflowControllerCookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
	ClearCookie(def *httputil.CookieDef) *http.Cookie
}

type AuthflowControllerSessionStore interface {
	Get(ctx context.Context, id string) (*webapp.Session, error)
	Create(ctx context.Context, session *webapp.Session) (err error)
	Update(ctx context.Context, session *webapp.Session) (err error)
	Delete(ctx context.Context, id string) (err error)
}

type AuthflowControllerAuthflowService interface {
	CreateNewFlow(ctx context.Context, intent authflow.PublicFlow, sessionOptions *authflow.SessionOptions) (*authflow.ServiceOutput, error)
	Get(ctx context.Context, stateToken string) (*authflow.ServiceOutput, error)
	FeedInput(ctx context.Context, stateToken string, rawMessage json.RawMessage) (*authflow.ServiceOutput, error)
}

type AuthflowControllerOAuthSessionService interface {
	Get(ctx context.Context, entryID string) (*oauthsession.Entry, error)
}

type AuthflowControllerSAMLSessionService interface {
	Get(ctx context.Context, entryID string) (*samlsession.SAMLSession, error)
}

type AuthflowControllerUIInfoResolver interface {
	ResolveForUI(ctx context.Context, r protocol.AuthorizationRequest) (*oidc.UIInfo, error)
	LooksLikeOriginallyTriggeredByOIDCOrSaml(r *http.Request) bool
}

type AuthflowControllerOAuthClientResolver interface {
	ResolveClient(clientID string) *config.OAuthClientConfig
}

type AuthflowNavigator interface {
	Navigate(ctx context.Context, screen *webapp.AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *webapp.Result)
	NavigateSelectAccount(result *webapp.Result)
	NavigateResetPasswordSuccessPage() string
	NavigateVerifyBotProtection(result *webapp.Result)
	NavigateOAuthProviderDemoCredentialPage(screen *webapp.AuthflowScreen, r *http.Request) *webapp.Result
}

var AuthflowControllerLogger = slogutil.NewLogger("authflow_controller")

func GetXStepFromQuery(r *http.Request) string {
	xStep := r.URL.Query().Get(webapp.AuthflowQueryKey)
	return xStep
}

type AuthflowOAuthCallbackResponse struct {
	Query string
	State *webappoauth.WebappOAuthState
}

type AuthflowEndpoints interface {
	SSOCallbackURL(alias string) *url.URL
	SharedSSOCallbackURL() *url.URL
}

type AuthflowController struct {
	TesterEndpointsProvider tester.EndpointsProvider
	TrustProxy              config.TrustProxy
	Clock                   clock.Clock

	Cookies        AuthflowControllerCookieManager
	Sessions       AuthflowControllerSessionStore
	SessionCookie  webapp.SessionCookieDef
	SignedUpCookie webapp.SignedUpCookieDef
	Endpoints      AuthflowEndpoints

	Authflows AuthflowControllerAuthflowService

	OAuthSessions  AuthflowControllerOAuthSessionService
	SAMLSessions   AuthflowControllerSAMLSessionService
	UIInfoResolver AuthflowControllerUIInfoResolver

	UIConfig            *config.UIConfig
	OAuthClientResolver AuthflowControllerOAuthClientResolver

	Navigator     AuthflowNavigator
	ErrorRenderer *ErrorRenderer
}

func (c *AuthflowController) HandleStartOfFlow(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	opts webapp.SessionOptions,
	flowType authflow.FlowType,
	handlers *AuthflowControllerHandlers,
	input interface{}) {
	if handled := c.handleInlinePreviewIfNecessary(ctx, w, r, handlers); handled {
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s, err := c.getOrCreateWebSession(ctx, w, r, opts)
	if err != nil {
		c.renderError(ctx, w, r, err)
		return
	}

	handleWithScreen := func(screen *webapp.AuthflowScreenWithFlowResponse) {
		handler := c.makeHTTPHandler(s, screen, handlers)
		handler.ServeHTTP(w, r)
	}

	screen, err := c.GetScreen(ctx, s, GetXStepFromQuery(r))
	if err != nil {
		if errors.Is(err, authflow.ErrFlowNotFound) {
			screen, err := c.createScreen(ctx, r, s, flowType, input)
			if errors.Is(err, declarative.ErrNoPublicSignup) {
				c.renderError(ctx, w, r, err)
				return
			} else if err != nil {
				c.renderError(ctx, w, r, err)
				return
			}

			// We used to redirect to the same page with x_step added.
			// But that could be blocked by mobile Safari 17 private mode.
			// See https://github.com/authgear/authgear-server/issues/3470
			handleWithScreen(screen)
			return
		}

		c.renderError(ctx, w, r, err)
		return
	}

	err = c.checkPath(ctx, w, r, s, screen)
	if err != nil {
		c.renderError(ctx, w, r, err)
		return
	}

	handleWithScreen(screen)
}

func (c *AuthflowController) HandleOAuthCallback(ctx context.Context, w http.ResponseWriter, r *http.Request, callbackResponse AuthflowOAuthCallbackResponse) {
	state := callbackResponse.State

	s, err := c.Sessions.Get(ctx, state.WebSessionID)
	if err != nil {
		c.renderError(ctx, w, r, err)
		return
	}

	screen, err := c.GetScreen(ctx, s, state.XStep)
	if err != nil {
		c.renderError(ctx, w, r, err)
		return
	}

	input := map[string]interface{}{
		"query": callbackResponse.Query,
	}
	result, err := c.AdvanceWithInput(ctx, r, s, screen, input, nil)
	if err != nil {
		u, parseURLErr := url.Parse(state.ErrorRedirectURI)
		if parseURLErr != nil {
			panic(parseURLErr)
		}

		c.ErrorRenderer.MakeAuthflowErrorResult(ctx, w, r, *u, err).WriteResponse(w, r)
		return
	}

	result.WriteResponse(w, r)
	return
}

func (c *AuthflowController) HandleResumeOfFlow(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	opts webapp.SessionOptions,
	handlers *AuthflowControllerHandlers,
	input map[string]interface{},
	errorHandler *AuthflowControllerErrorHandler,
) {
	handleError := func(err error) {
		if errorHandler != nil {
			err = (*errorHandler)(ctx, w, r, err)
			if err != nil {
				c.renderError(ctx, w, r, err)
			}
		}
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s, err := c.getOrCreateWebSession(ctx, w, r, opts)
	if err != nil {
		handleError(err)
		return
	}

	output, err := c.feedInput(ctx, "", input)
	if err != nil {
		handleError(err)
		return
	}

	screen, err := c.createScreenWithOutput(ctx, r, s, output, "")
	if errors.Is(err, declarative.ErrNoPublicSignup) {
		handleError(err)
		return
	} else if err != nil {
		handleError(err)
		return
	}

	err = c.checkPath(ctx, w, r, s, screen)
	if err != nil {
		handleError(err)
		return
	}

	handler := c.makeHTTPHandler(s, screen, handlers)
	handler.ServeHTTP(w, r)
}

func (c *AuthflowController) HandleStep(ctx context.Context, w http.ResponseWriter, r *http.Request, handlers *AuthflowControllerHandlers) {
	if handled := c.handleInlinePreviewIfNecessary(ctx, w, r, handlers); handled {
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s, err := c.getWebSession(ctx)
	if err != nil {
		c.renderError(ctx, w, r, err)
		return
	}

	screen, err := c.GetScreen(ctx, s, GetXStepFromQuery(r))
	if err != nil {
		c.renderError(ctx, w, r, err)
		return
	}

	err = c.checkPath(ctx, w, r, s, screen)
	if err != nil {
		c.renderError(ctx, w, r, err)
		return
	}

	handler := c.makeHTTPHandler(s, screen, handlers)
	handler.ServeHTTP(w, r)
}

func (c *AuthflowController) HandleWithoutScreen(ctx context.Context, w http.ResponseWriter, r *http.Request, handlers *AuthflowControllerHandlers) {
	if handled := c.handleInlinePreviewIfNecessary(ctx, w, r, handlers); handled {
		return
	}

	var session *webapp.Session
	s, err := c.getWebSession(ctx)
	if err != nil {
		c.renderError(ctx, w, r, err)
		return
	} else {
		session = s
	}

	handler := c.makeHTTPHandler(session, nil, handlers)
	handler.ServeHTTP(w, r)
}

func (c *AuthflowController) HandleWithoutSession(ctx context.Context, w http.ResponseWriter, r *http.Request, handlers *AuthflowControllerHandlers) {
	if handled := c.handleInlinePreviewIfNecessary(ctx, w, r, handlers); handled {
		return
	}

	handler := c.makeHTTPHandler(nil, nil, handlers)
	handler.ServeHTTP(w, r)
}

func (c *AuthflowController) handleInlinePreviewIfNecessary(ctx context.Context, w http.ResponseWriter, r *http.Request, handlers *AuthflowControllerHandlers) bool {
	if webapp.IsInlinePreviewPageRequest(r) && handlers.InlinePreviewHandler != nil {
		if err := handlers.InlinePreviewHandler(ctx, w, r); err != nil {
			c.renderError(ctx, w, r, err)
		}
		return true
	}
	return false
}

func (c *AuthflowController) getWebSession(ctx context.Context) (*webapp.Session, error) {
	s := webapp.GetSession(ctx)
	if s == nil {
		return nil, webapp.ErrSessionNotFound
	}
	if s.IsCompleted {
		return nil, webapp.ErrSessionCompleted
	}
	return s, nil
}

func (c *AuthflowController) getOrCreateWebSession(ctx context.Context, w http.ResponseWriter, r *http.Request, opts webapp.SessionOptions) (*webapp.Session, error) {
	now := c.Clock.NowUTC()
	s, err := c.getWebSession(ctx)
	if err == nil && s != nil {
		return s, nil
	}

	// If any other error, error out.
	if !(apierrors.IsKind(err, webapp.WebUIInvalidSession) || apierrors.IsKind(err, webapp.WebUISessionCompleted)) {
		return nil, err
	}

	// So err is either WebUIInvalidSession or WebUISessionCompleted.
	// If err is WebUIInvalidSession, we simply create a new session.
	// If err is WebUISessionCompleted, then it is complicated.
	//
	// If the authflow was originally triggered by OAuth / SAML, AND the authui page was navigated back,
	// then we want to keep the completed session.
	// What end-user will see is that the expired page.
	//
	// If the authflow was originally triggered by OAuth / SAML, AND the authui page was visited directly with /login or /signup,
	// then we want to re-create a new session.
	//
	// Here we use a simple heuristic to differentiate the 2 cases.
	if c.UIInfoResolver.LooksLikeOriginallyTriggeredByOIDCOrSaml(r) && apierrors.IsKind(err, webapp.WebUISessionCompleted) {
		return nil, err
	}

	o := opts
	o.UpdatedAt = now

	s = webapp.NewSession(o)
	err = c.Sessions.Create(ctx, s)
	if err != nil {
		return nil, err
	}

	cookie := c.Cookies.ValueCookie(c.SessionCookie.Def, s.ID)
	httputil.UpdateCookie(w, cookie)

	return s, nil
}

func (c *AuthflowController) GetScreen(ctx context.Context, s *webapp.Session, xStep string) (*webapp.AuthflowScreenWithFlowResponse, error) {
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

	output, err := c.Authflows.Get(ctx, screen.StateToken.StateToken)
	if err != nil {
		return nil, err
	}
	flowResponse := output.ToFlowResponse()
	screenWithResponse.StateTokenFlowResponse = &flowResponse

	if screen.BranchStateToken != nil {
		output, err = c.Authflows.Get(ctx, screen.BranchStateToken.StateToken)
		if err != nil {
			return nil, err
		}

		flowResponse := output.ToFlowResponse()
		screenWithResponse.BranchStateTokenFlowResponse = &flowResponse
	}

	return screenWithResponse, nil
}

func (c *AuthflowController) createAuthflow(ctx context.Context, r *http.Request, s *webapp.Session, flowType authflow.FlowType) (*authflow.ServiceOutput, error) {
	flowName := "default"

	if s.OAuthSessionID != "" {
		var err error
		flowName, err = c.deriveFlowNameFromOAuthSession(ctx, s.OAuthSessionID, flowType)
		if err != nil {
			return nil, err
		}
	} else if s.SAMLSessionID != "" {
		var err error
		flowName, err = c.deriveFlowNameFromSAMLSession(ctx, s.SAMLSessionID, flowType)
		if err != nil {
			return nil, err
		}
	}

	flowReference := authflow.FlowReference{
		Name: flowName,
		Type: flowType,
	}

	flow, err := authflow.InstantiateFlow(flowReference, jsonpointer.T{})
	if err != nil {
		return nil, err
	}

	var sessionOptionsFromOAuthOrSAML *authflow.SessionOptions
	if s.OAuthSessionID != "" {
		sessionOptionsFromOAuthOrSAML, err = c.makeSessionOptionsFromOAuth(ctx, s.OAuthSessionID)
		if errors.Is(err, oauthsession.ErrNotFound) {
			// Ignore this error.
		} else if err != nil {
			return nil, err
		}
	} else if s.SAMLSessionID != "" {
		sessionOptionsFromOAuthOrSAML, err = c.makeSessionOptionsFromSAML(ctx, s.SAMLSessionID)
		if errors.Is(err, samlsession.ErrNotFound) {
			// Ignore this error.
		} else if err != nil {
			return nil, err
		}
	}

	sessionOptionsFromQuery := c.makeSessionOptionsFromQuery(r)

	// The query overrides the cookie.
	sessionOptions := sessionOptionsFromOAuthOrSAML.PartiallyMergeFrom(sessionOptionsFromQuery)

	output, err := c.Authflows.CreateNewFlow(ctx, flow, sessionOptions)
	if err != nil {
		return nil, err
	}

	return output, err
}

// ReplaceScreen is for switching flow.
func (c *AuthflowController) ReplaceScreen(ctx context.Context, r *http.Request, s *webapp.Session, flowType authflow.FlowType, input map[string]interface{}) (result *webapp.Result, err error) {
	var screen *webapp.AuthflowScreenWithFlowResponse
	result = &webapp.Result{}

	output, err := c.createAuthflow(ctx, r, s, flowType)
	if err != nil {
		return
	}

	flowResponse := output.ToFlowResponse()
	emptyXStep := ""
	var emptyInput map[string]interface{}
	screen = webapp.NewAuthflowScreenWithFlowResponse(&flowResponse, emptyXStep, emptyInput)
	s.RememberScreen(screen.Screen)

	output, screen, err = c.takeBranchRecursively(ctx, s, screen)
	if err != nil {
		return
	}

	now := c.Clock.NowUTC()
	s.UpdatedAt = now
	err = c.Sessions.Update(ctx, s)
	if err != nil {
		return
	}

	result, err = c.AdvanceWithInput(ctx, r, s, screen, input, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *AuthflowController) Restart(ctx context.Context, s *webapp.Session) (result *webapp.Result, err error) {
	s.Authflow = nil
	// To be safe, also clear any interaction.
	s.Steps = []webapp.SessionStep{}
	now := c.Clock.NowUTC()
	s.UpdatedAt = now
	err = c.Sessions.Update(ctx, s)
	if err != nil {
		return
	}

	result = &webapp.Result{
		RemoveQueries: setutil.Set[string]{
			"x_step": struct{}{},
		},
	}
	c.Navigator.NavigateSelectAccount(result)
	return
}

func (c *AuthflowController) createScreenWithOutput(
	ctx context.Context,
	r *http.Request,
	s *webapp.Session,
	output *authflow.ServiceOutput,
	prevXStep string,
) (*webapp.AuthflowScreenWithFlowResponse, error) {
	flowResponse := output.ToFlowResponse()
	var emptyInput map[string]interface{}
	screen := webapp.NewAuthflowScreenWithFlowResponse(&flowResponse, prevXStep, emptyInput)
	s.RememberScreen(screen.Screen)

	output, screen, err := c.takeBranchRecursively(ctx, s, screen)
	if err != nil {
		return nil, err
	}

	now := c.Clock.NowUTC()
	s.UpdatedAt = now
	err = c.Sessions.Update(ctx, s)
	if err != nil {
		return nil, err
	}

	return screen, nil
}

func (c *AuthflowController) createScreen(
	ctx context.Context,
	r *http.Request,
	s *webapp.Session,
	flowType authflow.FlowType,
	input interface{}) (screen *webapp.AuthflowScreenWithFlowResponse, err error) {
	output1, err := c.createAuthflow(ctx, r, s, flowType)
	if err != nil {
		return
	}

	screen, err = c.createScreenWithOutput(ctx, r, s, output1, "")
	if err != nil {
		return
	}

	if input != nil {
		output2, err := c.feedInput(ctx, output1.Flow.StateToken, input)
		if err != nil {
			return nil, err
		}
		screen, err = c.createScreenWithOutput(ctx, r, s, output2, screen.Screen.StateToken.XStep)
		if err != nil {
			return nil, err
		}
	}

	return
}

type AdvanceOptions struct {
	// Treat the next screen as the extend of the taken branch
	// Allowing alternative branches to be selected based on the current state
	InheritTakenBranchState bool
}

// AdvanceWithInputs is for feeding multiple inputs that would advance the flow.
func (c *AuthflowController) AdvanceWithInputs(
	ctx context.Context,
	r *http.Request,
	s *webapp.Session,
	screen0 *webapp.AuthflowScreenWithFlowResponse,
	inputs []map[string]interface{},
	options *AdvanceOptions,
) (*webapp.Result, error) {
	if options == nil {
		options = &AdvanceOptions{}
	}
	result := &webapp.Result{}

	prevXStep := screen0.Screen.StateToken.XStep
	currentScreen := screen0

	for _, input := range inputs {
		output1, err := c.feedInput(ctx, currentScreen.Screen.StateToken.StateToken, input)
		if err != nil {
			return nil, err
		}

		result.Cookies = append(result.Cookies, output1.Cookies...)
		flowResponse1 := output1.ToFlowResponse()

		screen1 := webapp.NewAuthflowScreenWithFlowResponse(&flowResponse1, prevXStep, input)
		if options.InheritTakenBranchState && screen0.BranchStateTokenFlowResponse != nil {
			screen1.InheritTakenBranchState(screen0)
		}
		s.RememberScreen(screen1.Screen)
		currentScreen = screen1

		output2, screen2, err := c.takeBranchRecursively(ctx, s, screen1)
		if err != nil {
			return nil, err
		}
		currentScreen = screen2

		if output2 != nil {
			result.Cookies = append(result.Cookies, output2.Cookies...)
		}
		prevXStep = screen2.Screen.StateToken.XStep

		_ = *screen2.StateTokenFlowResponse

		err = c.updateSession(ctx, r, s)
		if err != nil {
			return nil, err
		}
	}

	currentScreen.Navigate(ctx, c.Navigator, r, s.ID, result)

	return result, nil
}

func (c *AuthflowController) Finish(ctx context.Context, r *http.Request, s *webapp.Session) (*webapp.Result, error) {
	if s == nil {
		panic(fmt.Errorf("unexpected: session is missing in Finish"))
	}
	if s.Authflow == nil {
		return nil, fmt.Errorf("cannot finish authflow: session is not in a flow")
	}

	result := &webapp.Result{}
	screen, ok := s.Authflow.AllScreens[GetXStepFromQuery(r)]
	if !ok {
		return nil, authflow.ErrFlowNotFound
	}
	err := c.finishSession(ctx, r, s, result, screen.FinishedUIScreenData)
	if err != nil {
		return nil, err
	}
	return result, nil
}
func (c *AuthflowController) AdvanceToDelayedScreen(r *http.Request, s *webapp.Session) (*webapp.Result, error) {
	if s.Authflow == nil {
		return nil, authflow.ErrFlowNotFound
	}

	screen, ok := s.Authflow.AllScreens[GetXStepFromQuery(r)]
	if !ok {
		return nil, authflow.ErrFlowNotFound
	}

	return screen.DelayedUIScreenData.TargetResult, nil
}

func (c *AuthflowController) DelayScreen(ctx context.Context, r *http.Request,
	s *webapp.Session,
	sourceScreen *webapp.AuthflowScreen,
	targetResult *webapp.Result,
	screenFactory func(*webapp.AuthflowScreen) *webapp.AuthflowScreen,
) (*webapp.AuthflowScreen, error) {
	screen := webapp.NewAuthflowDelayedScreenWithResult(sourceScreen, targetResult)
	if screenFactory != nil {
		screen = screenFactory(screen)
	}
	s.RememberScreen(screen)
	err := c.updateSession(ctx, r, s)
	if err != nil {
		return nil, err
	}

	return screen, nil
}

func (c *AuthflowController) AdvanceDirectly(
	route string,
	screen *webapp.AuthflowScreenWithFlowResponse,
	query url.Values,
) *webapp.Result {
	result := &webapp.Result{}
	screen.AdvanceWithQuery(route, result, query)
	return result
}

// AdvanceWithInput is same as AdvanceWithInputs but only allow one input.
func (c *AuthflowController) AdvanceWithInput(
	ctx context.Context,
	r *http.Request,
	s *webapp.Session,
	screen0 *webapp.AuthflowScreenWithFlowResponse,
	input map[string]interface{},
	options *AdvanceOptions,
) (result *webapp.Result, err error) {
	return c.AdvanceWithInputs(ctx, r, s, screen0, []map[string]interface{}{input}, options)
}

// UpdateWithInput is for feeding an input that would just update the current node.
// One application is resend.
func (c *AuthflowController) UpdateWithInput(ctx context.Context, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse, input map[string]interface{}) (result *webapp.Result, err error) {
	result = &webapp.Result{}

	output, err := c.feedInput(ctx, screen.Screen.StateToken.StateToken, input)
	if err != nil {
		return
	}

	result.Cookies = append(result.Cookies, output.Cookies...)
	newF := output.ToFlowResponse()
	newScreen := webapp.UpdateAuthflowScreenWithFlowResponse(screen, &newF)
	s.RememberScreen(newScreen.Screen)

	now := c.Clock.NowUTC()
	s.UpdatedAt = now
	err = c.Sessions.Update(ctx, s)
	if err != nil {
		return
	}

	if output != nil {
		result.Cookies = append(result.Cookies, output.Cookies...)
	}

	newScreen.Navigate(ctx, c.Navigator, r, s.ID, result)
	return
}

func (c *AuthflowController) handleTakeBranchResultInput(
	ctx context.Context,
	s *webapp.Session,
	screen *webapp.AuthflowScreenWithFlowResponse,
	takeBranchResult webapp.TakeBranchResultInput,
) (*authflow.ServiceOutput, *webapp.AuthflowScreenWithFlowResponse, error) {
	retry := func(originalOutput *authflow.ServiceOutput, err error) (*authflow.ServiceOutput, error) {
		retryInput := (*takeBranchResult.OnRetry)(err)
		if retryInput == nil {
			return originalOutput, err
		}
		var retryErr error
		var output *authflow.ServiceOutput
		output, retryErr = c.feedInput(ctx, screen.Screen.StateToken.StateToken, retryInput)
		if retryErr != nil {
			return output, retryErr
		}
		return output, nil
	}

	output, err := c.feedInput(ctx, screen.Screen.StateToken.StateToken, takeBranchResult.Input)
	if takeBranchResult.TransformOutput != nil {
		output, err = (*takeBranchResult.TransformOutput)(ctx, output, err, webapp.TransformerDependencies{
			Authflows: c.Authflows,
			Clock:     c.Clock,
		})
	}
	if err != nil {
		if takeBranchResult.OnRetry == nil {
			return output, nil, err
		}
		var retryErr error
		output, retryErr = retry(output, err)
		if retryErr != nil {
			return output, nil, retryErr
		}
	}
	flowResponse := output.ToFlowResponse()
	newScreen := takeBranchResult.NewAuthflowScreenFull(&flowResponse, err)
	s.RememberScreen(newScreen.Screen)
	return output, newScreen, nil
}

func (c *AuthflowController) takeBranchRecursively(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (output *authflow.ServiceOutput, newScreen *webapp.AuthflowScreenWithFlowResponse, err error) {
	for screen.HasBranchToTake() {
		// Take the first branch, and first channel by default.
		var zeroIndex int
		var zeroChannel model.AuthenticatorOOBChannel
		takeBranchResult := screen.TakeBranch(&webapp.TakeBranchInput{
			Index:   zeroIndex,
			Channel: zeroChannel,
		}, &webapp.TakeBranchOptions{})

		switch takeBranchResult := takeBranchResult.(type) {
		// This taken branch does not require an input to select.
		case webapp.TakeBranchResultSimple:
			s.RememberScreen(takeBranchResult.Screen.Screen)
			screen = takeBranchResult.Screen
		// This taken branch require an input to select.
		case webapp.TakeBranchResultInput:
			output, screen, err = c.handleTakeBranchResultInput(ctx, s, screen, takeBranchResult)
			if err != nil {
				return
			}
		}
	}

	newScreen = screen
	return
}

func (c *AuthflowController) FeedInputWithoutNavigate(ctx context.Context, stateToken string, input interface{}) (*authflow.ServiceOutput, error) {
	return c.feedInput(ctx, stateToken, input)
}

func (c *AuthflowController) feedInput(ctx context.Context, stateToken string, input interface{}) (*authflow.ServiceOutput, error) {
	rawMessageBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	rawMessage := json.RawMessage(rawMessageBytes)
	output, err := c.Authflows.FeedInput(ctx, stateToken, rawMessage)
	if err != nil && !errors.Is(err, authflow.ErrEOF) {
		return nil, err
	}

	return output, nil
}

func (c *AuthflowController) deriveAuthflowFinishPath(flowType authflow.FlowType) string {
	switch flowType {
	case authflow.FlowTypeAccountRecovery:
		return c.Navigator.NavigateResetPasswordSuccessPage()
	}
	return ""
}

func (c *AuthflowController) deriveFinishRedirectURI(r *http.Request, s *webapp.Session, flowType authflow.FlowType, finishRedirectURI string) string {
	// 1. Use predefined redirect path of the flow
	// 2. Use the finish_redirect_uri from authflow. (To return to /oauth2/consent)
	// 3. Use redirect URI in webapp.Session.
	// 4. DerivePostLoginRedirectURIFromRequest

	path := c.deriveAuthflowFinishPath(flowType)
	if path != "" {
		return path
	}

	if finishRedirectURI != "" {
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

func (c *AuthflowController) makeSessionOptionsFromOAuth(ctx context.Context, oauthSessionID string) (*authflow.SessionOptions, error) {
	entry, err := c.OAuthSessions.Get(ctx, oauthSessionID)
	if err != nil {
		return nil, err
	}
	req := entry.T.AuthorizationRequest

	uiInfo, err := c.UIInfoResolver.ResolveForUI(ctx, req)
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

func (c *AuthflowController) makeSessionOptionsFromSAML(ctx context.Context, samlSessionID string) (*authflow.SessionOptions, error) {
	entry, err := c.SAMLSessions.Get(ctx, samlSessionID)
	if err != nil {
		return nil, err
	}
	uiInfo := entry.UIInfo

	sessionOptions := &authflow.SessionOptions{
		SAMLSessionID: samlSessionID,

		ClientID:    uiInfo.SAMLServiceProviderID,
		RedirectURI: uiInfo.RedirectURI,
		Prompt:      uiInfo.Prompt,
		LoginHint:   uiInfo.LoginHint,
	}

	return sessionOptions, nil
}

func (c *AuthflowController) deriveFlowNameFromOAuthSession(ctx context.Context, oauthSessionID string, flowType authflow.FlowType) (string, error) {
	entry, err := c.OAuthSessions.Get(ctx, oauthSessionID)
	if err != nil {
		if errors.Is(err, oauthsession.ErrNotFound) {
			return "", errors.Join(webapp.ErrInvalidSession, err)
		} else {
			return "", err
		}
	}

	req := entry.T.AuthorizationRequest
	specifiedFlowGroup := req.AuthenticationFlowGroup()
	client := c.OAuthClientResolver.ResolveClient(req.ClientID())

	return DeriveFlowName(
		flowType,
		specifiedFlowGroup,
		client.AuthenticationFlowAllowlist,
		c.UIConfig.AuthenticationFlow.Groups,
	)
}

func (c *AuthflowController) deriveFlowNameFromSAMLSession(ctx context.Context, samlSessionID string, flowType authflow.FlowType) (string, error) {
	samlSession, err := c.SAMLSessions.Get(ctx, samlSessionID)
	if err != nil {
		if errors.Is(err, samlsession.ErrNotFound) {
			return "", errors.Join(webapp.ErrInvalidSession, err)
		} else {
			return "", err
		}
	}

	specifiedFlowGroup := "" // SAML cannot specify flow group in request
	client := c.OAuthClientResolver.ResolveClient(samlSession.Entry.ServiceProviderID)

	// TODO(DEV-2004): Remove the check after service provider `id` is completely removed
	var allowList *config.AuthenticationFlowAllowlist
	if client != nil {
		allowList = client.AuthenticationFlowAllowlist
	}

	return DeriveFlowName(
		flowType,
		specifiedFlowGroup,
		allowList,
		c.UIConfig.AuthenticationFlow.Groups,
	)
}

func DeriveFlowName(flowType authflow.FlowType, flowGroup string, clientAllowlist *config.AuthenticationFlowAllowlist, definedGroups []*config.UIAuthenticationFlowGroup) (string, error) {
	allowlist := authflow.NewFlowAllowlist(clientAllowlist, definedGroups)
	return allowlist.DeriveFlowNameForDefaultUI(flowType, flowGroup)
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
			err = handlers.GetHandler(r.Context(), s, screen)
		case http.MethodPost:
			xAction := r.FormValue("x_action")
			switch xAction {
			case "take_branch":
				err = c.takeBranch(r.Context(), w, r, s, screen)
			default:
				handler, ok := handlers.PostHandlers[xAction]
				if !ok {
					http.Error(w, "Unknown action", http.StatusBadRequest)
					return
				}

				err = handler(r.Context(), s, screen)
			}
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if err != nil {
			c.renderError(r.Context(), w, r, err)
		}
	})
}

func (c *AuthflowController) takeBranch(ctx context.Context, w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
	xStepAtBranch := screen.Screen.BranchStateToken.XStep
	screen, err := c.GetScreen(ctx, s, xStepAtBranch)
	if err != nil {
		return err
	}

	indexStr := r.Form.Get("x_index")
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return err
	}
	channel := r.Form.Get("x_channel")
	input := &webapp.TakeBranchInput{
		Index:   index,
		Channel: model.AuthenticatorOOBChannel(channel),
	}
	if hasBPInput := IsBotProtectionInputValid(ctx, r.Form); hasBPInput {
		input.BotProtectionProviderType = r.Form.Get("x_bot_protection_provider_type")
		input.BotProtectionProviderResponse = r.Form.Get("x_bot_protection_provider_response")
	}
	takeBranchResult := screen.TakeBranch(input, &webapp.TakeBranchOptions{
		DisableFallbackToSMS: true,
	})

	result := &webapp.Result{}

	var output *authflow.ServiceOutput
	var newScreen *webapp.AuthflowScreenWithFlowResponse
	switch takeBranchResult := takeBranchResult.(type) {
	// This taken branch does not require an input to select.
	case webapp.TakeBranchResultSimple:
		s.RememberScreen(takeBranchResult.Screen.Screen)
		newScreen = takeBranchResult.Screen
	// This taken branch require an input to select.
	case webapp.TakeBranchResultInput:
		output, newScreen, err = c.handleTakeBranchResultInput(ctx, s, screen, takeBranchResult)
		if err != nil {
			return err
		}
		result.Cookies = append(result.Cookies, output.Cookies...)
	}

	output2, newScreen, err := c.takeBranchRecursively(ctx, s, newScreen)
	if err != nil {
		return err
	}
	if output2 != nil {
		result.Cookies = append(result.Cookies, output2.Cookies...)
	}

	err = c.updateSession(ctx, r, s)
	if err != nil {
		return err
	}

	newScreen.Navigate(ctx, c.Navigator, r, s.ID, result)
	result.WriteResponse(w, r)
	return nil
}

func (c *AuthflowController) renderError(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	c.ErrorRenderer.RenderError(ctx, w, r, err)
}

func (c *AuthflowController) checkPath(ctx context.Context, w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
	logger := AuthflowControllerLogger.GetLogger(ctx)
	if screen.Screen != nil && screen.Screen.SkipPathCheck {
		return nil
	}

	// We derive the intended path of the screen,
	// and check if the paths match.
	result := &webapp.Result{}
	screen.Navigate(ctx, c.Navigator, r, s.ID, result)
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
		logger.Warn(ctx, "path mismatch")
		return webapp.ErrInvalidSession
	}

	return nil
}

func (c *AuthflowController) updateSession(
	ctx context.Context,
	r *http.Request,
	s *webapp.Session) (err error) {
	now := c.Clock.NowUTC()
	s.UpdatedAt = now
	return c.Sessions.Update(ctx, s)
}

func (c *AuthflowController) finishSession(
	ctx context.Context,
	r *http.Request,
	s *webapp.Session,
	result *webapp.Result,
	finishedUIScreenData *webapp.AuthflowFinishedUIScreenData,
) (err error) {
	result.RemoveQueries = setutil.Set[string]{
		"x_step": struct{}{},
	}
	result.NavigationAction = webapp.NavigationActionRedirect
	result.RedirectURI = c.deriveFinishRedirectURI(r, s, finishedUIScreenData.FlowType, finishedUIScreenData.FinishRedirectURI)

	switch finishedUIScreenData.FlowType {
	case authflow.FlowTypeLogin:
		fallthrough
	case authflow.FlowTypePromote:
		fallthrough
	case authflow.FlowTypeSignup:
		fallthrough
	case authflow.FlowTypeSignupLogin:
		fallthrough
	case authflow.FlowTypeReauth:
		// Mark the current session as completed,
		// so that user will see a completed screen if trying to back to the previous steps
		s.IsCompleted = true
		err = c.Sessions.Update(ctx, s)
		if err != nil {
			return err
		}
		// Marked signed up in cookie after authorization.
		// When user visit auth ui root "/", redirect user to "/login" if
		// cookie exists
		result.Cookies = append(result.Cookies, c.Cookies.ValueCookie(c.SignedUpCookie.Def, "true"))
		// Reset visitor ID.
		result.Cookies = append(result.Cookies, c.Cookies.ClearCookie(webapp.VisitorIDCookieDef))
	default:
		// Do nothing for other flows
	}
	return nil
}

func (c *AuthflowController) GetAccountLinkingSSOCallbackURL(alias string, data declarative.AccountLinkingIdentifyData) (string, error) {
	var idenOption *declarative.AccountLinkingIdentificationOption
	for _, option := range data.Options {
		option := option
		if option.Alias == alias {
			idenOption = &option
			break
		}
	}
	if idenOption == nil {
		return "", fmt.Errorf("unknown alias %s", alias)
	}

	var callbackURL string
	if idenOption.ProviderStatus == config.OAuthProviderStatusUsingDemoCredentials {
		callbackURL = c.Endpoints.SharedSSOCallbackURL().String()
	} else {
		callbackURL = c.Endpoints.SSOCallbackURL(alias).String()
	}
	return callbackURL, nil
}

func (c *AuthflowController) UseOAuthIdentification(
	ctx context.Context, s *webapp.Session,
	w http.ResponseWriter, r *http.Request,
	screen *webapp.AuthflowScreen,
	alias string, identificationOptions []declarative.IdentificationOption,
	flowExecutor func(input map[string]interface{},
	) (
		result *webapp.Result,
		err error),
) (*webapp.Result, error) {
	var idenOption *declarative.IdentificationOption
	for _, option := range identificationOptions {
		option := option
		if option.Alias == alias {
			idenOption = &option
			break
		}
	}
	if idenOption == nil {
		return nil, fmt.Errorf("unknown alias %s", alias)
	}
	var callbackURL string
	if idenOption.ProviderStatus == config.OAuthProviderStatusUsingDemoCredentials {
		callbackURL = c.Endpoints.SharedSSOCallbackURL().String()
	} else {
		callbackURL = c.Endpoints.SSOCallbackURL(alias).String()
	}

	input := map[string]interface{}{
		"identification": "oauth",
		"alias":          alias,
		"redirect_uri":   callbackURL,
		"response_mode":  oauthrelyingparty.ResponseModeFormPost,
	}

	result, err := flowExecutor(input)
	if err != nil {
		return nil, err
	}

	if idenOption.ProviderStatus == config.OAuthProviderStatusUsingDemoCredentials {
		// Insert a confirmation screen if the provider is using demo credentials
		newScreen, err := c.DelayScreen(ctx, r, s, screen, result, func(newScreen *webapp.AuthflowScreen) *webapp.AuthflowScreen {
			newScreen.OAuthProviderDemoCredentialViewModel = &authflowv2viewmodels.OAuthProviderDemoCredentialViewModel{
				ProviderType: idenOption.ProviderType,
				FromURL:      r.URL.String(),
			}
			return newScreen
		})
		if err != nil {
			return nil, err
		}
		newResult := c.Navigator.NavigateOAuthProviderDemoCredentialPage(newScreen, r)
		result = newResult
	}

	return result, nil
}
