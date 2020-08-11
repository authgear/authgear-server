package webapp

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/intents"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
	"github.com/authgear/authgear-server/pkg/core/base32"
	corerand "github.com/authgear/authgear-server/pkg/core/rand"
	"github.com/authgear/authgear-server/pkg/core/skyerr"
	"github.com/authgear/authgear-server/pkg/httputil"
	"github.com/authgear/authgear-server/pkg/log"
)

//go:generate mockgen -source=service.go -destination=service_mock_test.go -package webapp

type RedirectURIGetter interface {
	GetRedirectURI() string
}

type ErrorRedirectURIGetter interface {
	GetErrorRedirectURI() string
}

type CookiesGetter interface {
	GetCookies() []*http.Cookie
}

type Store interface {
	Create(state *State) error
	Get(instanceID string) (*State, error)
}

type GraphService interface {
	NewGraph(ctx *newinteraction.Context, intent newinteraction.Intent) (*newinteraction.Graph, error)
	Get(instanceID string) (*newinteraction.Graph, error)
	DryRun(webStateID string, fn func(*newinteraction.Context) (*newinteraction.Graph, error)) error
	Run(webStateID string, graph *newinteraction.Graph, preserveGraph bool) error
}

type CookieFactory interface {
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
}

type ServiceLogger struct{ *log.Logger }

func NewServiceLogger(lf *log.Factory) ServiceLogger {
	return ServiceLogger{lf.New("webapp-service")}
}

var UserAgentTokenCookie = &httputil.CookieDef{
	Name:              "ua-token",
	Path:              "/",
	AllowScriptAccess: false,
	SameSite:          http.SameSiteNoneMode, // Ensure resume-able after redirecting from external site
	MaxAge:            nil,                   // Use HTTP session cookie; expires when browser closes
}

type Service struct {
	Logger        ServiceLogger
	Request       *http.Request
	Store         Store
	Graph         GraphService
	CookieFactory CookieFactory
}

func (s *Service) getUserAgentToken() string {
	token, err := s.Request.Cookie(UserAgentTokenCookie.Name)
	if err != nil {
		return ""
	}
	return token.Value
}

func (s *Service) generateUserAgentToken() (string, *http.Cookie) {
	token := corerand.StringWithAlphabet(32, base32.Alphabet, corerand.SecureRand)
	cookie := s.CookieFactory.ValueCookie(UserAgentTokenCookie, token)
	return token, cookie
}

func (s *Service) GetState(stateID string) (*State, error) {
	if stateID == "" {
		return nil, nil
	}

	state, err := s.Store.Get(stateID)
	if err != nil {
		return nil, err
	}
	if state.UserAgentToken != s.getUserAgentToken() {
		return nil, ErrInvalidState
	}
	return state, nil
}

func (s *Service) GetIntent(intent *Intent, stateID string) (state *State, graph *newinteraction.Graph, err error) {
	var stateError *skyerr.APIError
	if stateID != "" {
		state, err = s.GetState(stateID)
		if err != nil {
			return
		}

		// Ignore the intent, reuse the existing state.
		newID := intent.StateID
		if newID == "" {
			newID = NewID()
		}
		state.SetID(newID)
		graph, err = s.get(state)
		if err == nil || !errors.Is(err, newinteraction.ErrStateNotFound) {
			return
		}

		// If graph is not found: use the intent, with same state error
		stateError = state.Error
	}

	newStateID := intent.StateID
	if newStateID == "" {
		newStateID = NewID()
	}

	err = s.Graph.DryRun(newStateID, func(ctx *newinteraction.Context) (*newinteraction.Graph, error) {
		graph, err = s.Graph.NewGraph(ctx, intent.Intent)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return
	}

	state = &State{
		ID:              newStateID,
		RedirectURI:     intent.RedirectURI,
		GraphInstanceID: graph.InstanceID,
		Extra:           map[string]interface{}{},
		Error:           stateError,
	}

	return
}

func (s *Service) Get(stateID string) (state *State, graph *newinteraction.Graph, err error) {
	state, err = s.GetState(stateID)
	if err != nil {
		return
	}

	graph, err = s.get(state)
	return
}

func (s *Service) get(state *State) (graph *newinteraction.Graph, err error) {
	graph, err = s.Graph.Get(state.GraphInstanceID)
	if err != nil {
		return
	}

	err = s.Graph.DryRun(state.ID, func(ctx *newinteraction.Context) (*newinteraction.Graph, error) {
		err = graph.Apply(ctx)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return
	}

	return
}

func (s *Service) PostIntent(intent *Intent, inputer func() (interface{}, error)) (result *Result, err error) {
	stateID := intent.StateID
	if stateID != "" {
		// Ignore the intent, reuse the existing state.
		state, err := s.Store.Get(stateID)
		if err != nil {
			return nil, err
		}

		result, err = s.post(state, inputer)
		if err == nil || !errors.Is(err, newinteraction.ErrStateNotFound) {
			return result, err
		}

		// If graph is not found: use the intent
	}

	state := &State{
		ID:          NewID(),
		RedirectURI: intent.RedirectURI,
		KeepState:   intent.KeepState,
		UILocales:   intent.UILocales,
	}

	var graph *newinteraction.Graph
	var edges []newinteraction.Edge
	err = s.Graph.DryRun(state.ID, func(ctx *newinteraction.Context) (newGraph *newinteraction.Graph, err error) {
		graph, err = s.Graph.NewGraph(ctx, intent.Intent)
		if err != nil {
			return
		}

		input, err := inputer()
		if err != nil {
			return
		}

		graph, edges, err = graph.Accept(ctx, input)
		if errors.Is(err, newinteraction.ErrInputRequired) {
			err = nil
			newGraph = graph
			return
		}
		if err != nil {
			return
		}

		newGraph = graph
		return
	})

	// Regardless of err, we need to return result.
	result, err = s.afterPost(state, graph, edges, err)
	if err != nil {
		return
	}

	return
}

func (s *Service) PostInput(stateID string, inputer func() (interface{}, error)) (result *Result, err error) {
	state, err := s.GetState(stateID)
	if err != nil {
		return nil, err
	}

	return s.post(state, inputer)
}

func (s *Service) post(state *State, inputer func() (interface{}, error)) (result *Result, err error) {
	// Immutable state
	state.SetID(NewID())

	var edges []newinteraction.Edge
	graph, err := s.Graph.Get(state.GraphInstanceID)
	if err != nil {
		return nil, err
	}

	err = s.Graph.DryRun(state.ID, func(ctx *newinteraction.Context) (*newinteraction.Graph, error) {
		input, err := inputer()
		if err != nil {
			return nil, err
		}

		err = graph.Apply(ctx)
		if err != nil {
			return nil, err
		}

		var newGraph *newinteraction.Graph
		newGraph, edges, err = graph.Accept(ctx, input)
		if errors.Is(err, newinteraction.ErrInputRequired) {
			graph = newGraph
			return newGraph, nil
		} else if err != nil {
			return nil, err
		}

		graph = newGraph
		return newGraph, nil
	})

	// Regardless of err, we need to return result.
	result, err = s.afterPost(state, graph, edges, err)
	if err != nil {
		return
	}

	return
}

func (s *Service) afterPost(state *State, graph *newinteraction.Graph, edges []newinteraction.Edge, inputError error) (*Result, error) {
	finished := graph != nil && len(edges) == 0 && inputError == nil
	// The graph finished. Apply its effect permanently
	if finished {
		s.Logger.Debugf("afterPost run graph")
		inputError = s.Graph.Run(state.ID, graph, false)
	}

	var cookies []*http.Cookie

	state.UserAgentToken = s.getUserAgentToken()
	if state.UserAgentToken == "" {
		token, userAgentTokenCookie := s.generateUserAgentToken()
		state.UserAgentToken = token
		cookies = append(cookies, userAgentTokenCookie)
	}

	state.Error = skyerr.AsAPIError(inputError)
	if graph != nil {
		state.GraphInstanceID = graph.InstanceID
		state.Extra = s.collectExtras(graph.CurrentNode())
	}

	err := s.Store.Create(state)
	if err != nil {
		return nil, err
	}

	// Case: error
	if state.Error != nil {
		if !skyerr.IsAPIError(inputError) {
			s.Logger.Errorf("afterPost error: %v", inputError)
		}
		if graph != nil {
			node := graph.CurrentNode()
			if a, ok := node.(ErrorRedirectURIGetter); ok {
				errorRedirectURIString := a.GetErrorRedirectURI()
				errorRedirectURI, err := url.Parse(errorRedirectURIString)
				if err != nil {
					return nil, err
				}
				return &Result{
					state:            state,
					errorRedirectURI: errorRedirectURI,
					cookies:          cookies,
				}, nil
			}
		}
		return &Result{
			state:   state,
			cookies: cookies,
		}, nil
	}

	// Case: finish
	if finished {
		s.Logger.Debugf("afterPost finished")
		// Loop from start to end to collect cookies.
		// This iteration order allows newer node to overwrite cookies.
		for _, node := range graph.Nodes {
			if a, ok := node.(CookiesGetter); ok {
				cookies = append(cookies, a.GetCookies()...)
			}
		}

		redirectURI, err := url.Parse(state.RedirectURI)
		if err != nil {
			return nil, err
		}

		if state.KeepState {
			redirectURI = AttachStateID(state.ID, redirectURI)
		}

		return &Result{
			state:       state,
			redirectURI: redirectURI.String(),
			cookies:     cookies,
		}, nil
	}

	s.Logger.Debugf("afterPost transition")

	node := graph.CurrentNode()

	// Case: the node has redirect URI.
	if a, ok := node.(RedirectURIGetter); ok {
		s.Logger.Debugf("afterPost transition to node redirect URI")
		return &Result{
			state:       state,
			redirectURI: a.GetRedirectURI(),
			cookies:     cookies,
		}, nil
	}

	// Case: transition
	path, statusCode := s.deriveRedirectPath(graph)

	s.Logger.Debugf("afterPost transition to path: %v", path)

	pathURL, err := url.Parse(path)
	if err != nil {
		panic(fmt.Errorf("webapp: unexpected invalid transition path: %v", path))
	}
	redirectURI := AttachStateID(state.ID, pathURL).String()

	s.Logger.Debugf("afterPost transition to redirect URI: %v", redirectURI)

	return &Result{
		state:       state,
		redirectURI: redirectURI,
		statusCode:  statusCode,
		cookies:     cookies,
	}, nil
}

// nolint:gocyclo
func (s *Service) deriveRedirectPath(graph *newinteraction.Graph) (path string, statusCode int) {
	statusCode = http.StatusFound

	switch graph.CurrentNode().(type) {
	case *nodes.NodeSelectIdentityBegin:
		switch intent := graph.Intent.(type) {
		case *intents.IntentAuthenticate:
			switch intent.Kind {
			case intents.IntentAuthenticateKindLogin:
				path = "/login"
			case intents.IntentAuthenticateKindSignup:
				path = "/signup"
			default:
				panic(fmt.Errorf("webapp: unexpected authenticate intent: %T", intent.Kind))
			}
		default:
			panic(fmt.Errorf("webapp: unexpected intent: %T", graph.Intent))
		}
	case *nodes.NodeDoUseUser:
		switch graph.Intent.(type) {
		case *intents.IntentRemoveIdentity:
			path = "/enter_login_id"
		default:
			panic(fmt.Errorf("webapp: unexpected intent: %T", graph.Intent))
		}
	case *nodes.NodeUpdateIdentityBegin:
		path = "/enter_login_id"
	case *nodes.NodeCreateIdentityBegin:
		switch intent := graph.Intent.(type) {
		case *intents.IntentAuthenticate:
			switch intent.Kind {
			case intents.IntentAuthenticateKindPromote:
				path = "/promote_user"
			default:
				panic(fmt.Errorf("webapp: unexpected authenticate intent: %T", intent.Kind))
			}
		case *intents.IntentAddIdentity:
			path = "/enter_login_id"
		default:
			panic(fmt.Errorf("webapp: unexpected intent: %T", graph.Intent))
		}
	case *nodes.NodeAuthenticationBegin:
		path = "/authentication_begin"
		statusCode = http.StatusTemporaryRedirect
	case *nodes.NodeCreateAuthenticatorBegin:
		path = "/create_authenticator_begin"
		statusCode = http.StatusTemporaryRedirect
	case *nodes.NodeAuthenticationOOBTrigger:
		path = "/enter_oob_otp"
	case *nodes.NodeCreateAuthenticatorOOBSetup:
		path = "/enter_oob_otp"
	case *nodes.NodeCreateAuthenticatorTOTPSetup:
		path = "/setup_totp"
	case *nodes.NodeGenerateRecoveryCodeBegin:
		path = "/setup_recovery_code"
	case *nodes.NodeVerifyIdentity:
		path = "/verify_identity"
	default:
		panic(fmt.Errorf("webapp: unexpected node: %T", graph.CurrentNode()))
	}

	return
}

func (s *Service) collectExtras(node newinteraction.Node) map[string]interface{} {
	switch node := node.(type) {
	case *nodes.NodeForgotPasswordEnd:
		return map[string]interface{}{
			"login_id": node.LoginID,
		}
	case *nodes.NodeEnsureVerificationEnd:
		return map[string]interface{}{
			"display_id": node.Identity.DisplayID(),
		}
	case *nodes.NodeDoVerifyIdentity:
		return map[string]interface{}{
			"display_id": node.Identity.DisplayID(),
		}
	default:
		return nil
	}
}
