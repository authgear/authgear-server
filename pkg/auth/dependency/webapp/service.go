package webapp

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/intents"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
	"github.com/authgear/authgear-server/pkg/core/skyerr"
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

type ServiceLogger struct{ *log.Logger }

func NewServiceLogger(lf *log.Factory) ServiceLogger {
	return ServiceLogger{lf.New("webapp-service")}
}

type Service struct {
	Logger ServiceLogger
	Store  Store
	Graph  GraphService
}

func (s *Service) GetState(stateID string) (state *State, err error) {
	if stateID != "" {
		state, err = s.Store.Get(stateID)
		if err != nil {
			return
		}
	}
	return
}

func (s *Service) GetIntent(intent *Intent, stateID string) (state *State, graph *newinteraction.Graph, edges []newinteraction.Edge, err error) {
	var stateError *skyerr.APIError
	if stateID != "" {
		state, err = s.Store.Get(stateID)
		if err != nil {
			return
		}
		stateError = state.Error
	}

	newStateID := intent.StateID
	if newStateID == "" {
		newStateID = NewID()
	}

	err = s.Graph.DryRun(newStateID, func(ctx *newinteraction.Context) (_ *newinteraction.Graph, err error) {
		graph, err = s.Graph.NewGraph(ctx, intent.Intent)
		if err != nil {
			return
		}

		node := graph.CurrentNode()
		edges, err = node.DeriveEdges(ctx, graph)
		if err != nil {
			return
		}

		return
	})
	if err != nil {
		return
	}

	state = &State{
		ID:              newStateID,
		RedirectURI:     intent.RedirectURI,
		GraphInstanceID: graph.InstanceID,
		Error:           stateError,
	}

	return
}

func (s *Service) Get(stateID string) (state *State, graph *newinteraction.Graph, edges []newinteraction.Edge, err error) {
	state, err = s.Store.Get(stateID)
	if err != nil {
		return
	}

	graph, err = s.Graph.Get(state.GraphInstanceID)
	if err != nil {
		return
	}

	err = s.Graph.DryRun(stateID, func(ctx *newinteraction.Context) (_ *newinteraction.Graph, err error) {
		err = graph.Apply(ctx)
		if err != nil {
			return nil, err
		}

		node := graph.CurrentNode()
		edges, err = node.DeriveEdges(ctx, graph)
		if err != nil {
			return
		}

		return
	})
	if err != nil {
		return
	}

	return
}

func (s *Service) PostIntent(intent *Intent, inputer func() (interface{}, error)) (result *Result, err error) {
	stateID := intent.StateID
	if stateID == "" {
		stateID = NewID()
	}
	state := &State{
		ID:          stateID,
		RedirectURI: intent.RedirectURI,
		KeepState:   intent.KeepState,
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
	state, err := s.Store.Get(stateID)
	if err != nil {
		return nil, err
	}

	// Immutable state
	state.ID = NewID()

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
				}, nil
			}
		}
		return &Result{
			state: state,
		}, nil
	}

	// Case: finish
	if finished {
		s.Logger.Debugf("afterPost finished")
		// Loop from start to end to collect cookies.
		// This iteration order allows newer node to overwrite cookies.
		var cookies []*http.Cookie
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
			redirectURI = state.Attach(redirectURI)
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
		}, nil
	}

	// Case: transition
	path := s.deriveRedirectPath(graph, edges)

	s.Logger.Debugf("afterPost transition to path: %v", path)

	pathURL, err := url.Parse(path)
	if err != nil {
		panic(fmt.Errorf("webapp: unexpected invalid transition path: %v", path))
	}
	redirectURI := state.Attach(pathURL).String()

	s.Logger.Debugf("afterPost transition to redirect URI: %v", redirectURI)

	return &Result{
		state:       state,
		redirectURI: redirectURI,
	}, nil
}

func (s *Service) deriveRedirectPath(graph *newinteraction.Graph, edges []newinteraction.Edge) string {
	firstEdge := edges[0]

	switch graph.CurrentNode().(type) {
	case *nodes.NodeDoUseUser:
		switch graph.Intent.(type) {
		case *intents.IntentRemoveIdentity:
			return "/enter_login_id"
		default:
			panic(fmt.Errorf("webapp: unexpected intent: %T", graph.Intent))
		}
	case *nodes.NodeUpdateIdentityBegin:
		return "/enter_login_id"
	case *nodes.NodeCreateIdentityBegin:
		switch intent := graph.Intent.(type) {
		case *intents.IntentAuthenticate:
			switch intent.Kind {
			case intents.IntentAuthenticateKindPromote:
				return "/promote_user"
			default:
				panic(fmt.Errorf("webapp: unexpected authenticate intent: %T", intent.Kind))
			}
		case *intents.IntentAddIdentity:
			return "/enter_login_id"
		default:
			panic(fmt.Errorf("webapp: unexpected intent: %T", graph.Intent))
		}
	case *nodes.NodeAuthenticationBegin:
		switch firstEdge.(type) {
		case *nodes.EdgeAuthenticationPassword:
			return "/enter_password"
		default:
			panic(fmt.Errorf("webapp: unexpected edge: %T", firstEdge))
		}
	case *nodes.NodeCreateAuthenticatorBegin:
		switch firstEdge.(type) {
		case *nodes.EdgeCreateAuthenticatorPassword:
			return "/create_password"
		case *nodes.EdgeCreateAuthenticatorOOBSetup:
			return "/setup_oob_otp"
		default:
			panic(fmt.Errorf("webapp: unexpected edge: %T", firstEdge))
		}
	case *nodes.NodeAuthenticationOOBTrigger:
		return "/enter_oob_otp"
	case *nodes.NodeCreateAuthenticatorOOBSetup:
		return "/enter_oob_otp"
	case *nodes.NodeVerifyIdentity:
		return "/verify_identity"
	default:
		panic(fmt.Errorf("webapp: unexpected node: %T", graph.CurrentNode()))
	}
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
