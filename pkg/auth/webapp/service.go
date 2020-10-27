package webapp

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	corerand "github.com/authgear/authgear-server/pkg/util/rand"
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
	NewGraph(ctx *interaction.Context, intent interaction.Intent) (*interaction.Graph, error)
	Get(instanceID string) (*interaction.Graph, error)
	DryRun(webStateID string, fn func(*interaction.Context) (*interaction.Graph, error)) error
	Run(webStateID string, graph *interaction.Graph) error
}

type CookieFactory interface {
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
}

type ServiceLogger struct{ *log.Logger }

func NewServiceLogger(lf *log.Factory) ServiceLogger {
	return ServiceLogger{lf.New("webapp-service")}
}

type Service struct {
	Logger        ServiceLogger
	Request       *http.Request
	Store         Store
	Graph         GraphService
	CookieFactory CookieFactory
	UATokenCookie CookieDef
}

func (s *Service) getUserAgentToken() string {
	token, err := s.Request.Cookie(s.UATokenCookie.Def.Name)
	if err != nil {
		return ""
	}
	return token.Value
}

func (s *Service) generateUserAgentToken() (string, *http.Cookie) {
	token := corerand.StringWithAlphabet(32, base32.Alphabet, corerand.SecureRand)
	cookie := s.CookieFactory.ValueCookie(s.UATokenCookie.Def, token)
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

func (s *Service) GetIntent(intent *Intent) (state *State, graph *interaction.Graph, err error) {
	state = NewState(intent)
	if intent.OldStateID != "" {
		var oldState *State
		oldState, err = s.GetState(intent.OldStateID)
		if err != nil {
			return
		}

		state.RestoreFrom(oldState)
	}

	err = s.Graph.DryRun(state.ID, func(ctx *interaction.Context) (*interaction.Graph, error) {
		graph, err = s.Graph.NewGraph(ctx, intent.Intent)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return
	}

	state.GraphInstanceID = graph.InstanceID

	return
}

func (s *Service) Get(stateID string) (state *State, graph *interaction.Graph, err error) {
	state, err = s.GetState(stateID)
	if err != nil {
		return
	}

	graph, err = s.get(state)
	return
}

func (s *Service) get(state *State) (graph *interaction.Graph, err error) {
	graph, err = s.Graph.Get(state.GraphInstanceID)
	if err != nil {
		return
	}

	err = s.Graph.DryRun(state.ID, func(ctx *interaction.Context) (*interaction.Graph, error) {
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
	state := NewState(intent)

	if intent.OldStateID != "" {
		// Ignore the intent, reuse the existing state.
		oldState, err := s.Store.Get(intent.OldStateID)
		if err != nil {
			return nil, err
		}

		state.RestoreFrom(oldState)
	}

	var graph *interaction.Graph
	var edges []interaction.Edge
	var clearCookie *interaction.ErrClearCookie
	err = s.Graph.DryRun(state.ID, func(ctx *interaction.Context) (newGraph *interaction.Graph, err error) {
		graph, err = s.Graph.NewGraph(ctx, intent.Intent)
		if err != nil {
			return
		}

		input, err := inputer()
		if err != nil {
			return
		}

		graph, edges, err = graph.Accept(ctx, input)

		errors.As(err, &clearCookie)

		var inputRequired *interaction.ErrInputRequired
		if errors.As(err, &inputRequired) {
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
	result, err = s.afterPost(state, graph, edges, err, clearCookie, true)
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

	// Immutable state
	state.SetID(NewID())

	var edges []interaction.Edge
	graph, err := s.Graph.Get(state.GraphInstanceID)
	if err != nil {
		return nil, err
	}

	var clearCookie *interaction.ErrClearCookie
	err = s.Graph.DryRun(state.ID, func(ctx *interaction.Context) (*interaction.Graph, error) {
		input, err := inputer()
		if err != nil {
			return nil, err
		}

		err = graph.Apply(ctx)
		if err != nil {
			return nil, err
		}

		var newGraph *interaction.Graph
		newGraph, edges, err = graph.Accept(ctx, input)

		errors.As(err, &clearCookie)

		var inputRequired *interaction.ErrInputRequired
		if errors.As(err, &inputRequired) {
			graph = newGraph
			return newGraph, nil
		} else if err != nil {
			return nil, err
		}

		graph = newGraph
		return newGraph, nil
	})

	// Regardless of err, we need to return result.
	result, err = s.afterPost(state, graph, edges, err, clearCookie, false)
	if err != nil {
		return
	}

	return
}

func (s *Service) afterPost(state *State, graph *interaction.Graph, edges []interaction.Edge, inputError error, clearCookie *interaction.ErrClearCookie, isNewIntent bool) (*Result, error) {
	finished := graph != nil && len(edges) == 0 && inputError == nil
	// The graph finished. Apply its effect permanently
	if finished {
		s.Logger.Debugf("afterPost run graph")
		inputError = s.Graph.Run(state.ID, graph)
	}

	var cookies []*http.Cookie

	if isNewIntent {
		token, userAgentTokenCookie := s.generateUserAgentToken()
		state.UserAgentToken = token
		cookies = append(cookies, userAgentTokenCookie)
	}

	// Handle ErrClearCookie specifically.
	// The main use case is to remove invalid device token.
	if clearCookie != nil {
		cookies = append(cookies, clearCookie.Cookies...)
	}

	state.Error = apierrors.AsAPIError(inputError)
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
		if !apierrors.IsAPIError(inputError) {
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
	path := s.deriveRedirectPath(graph)

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
		cookies:     cookies,
	}, nil
}

// nolint:gocyclo
func (s *Service) deriveRedirectPath(graph *interaction.Graph) (path string) {
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
		// Ideally we should use 307 but if we use 307
		// the CSRF middleware will complain about invalid CSRF token.
		path = "/authentication_begin"
	case *nodes.NodeCreateAuthenticatorBegin:
		path = "/create_authenticator_begin"
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

func (s *Service) collectExtras(node interaction.Node) map[string]interface{} {
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
