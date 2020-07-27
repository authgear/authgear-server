package webapp

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
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
	DryRun(fn func(*newinteraction.Context) (*newinteraction.Graph, error)) error
	Run(graph *newinteraction.Graph, preserveGraph bool) error
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

func (s *Service) GetIntent(intent *Intent, stateID string) (state *State, graph *newinteraction.Graph, edges []newinteraction.Edge, err error) {
	var stateError *skyerr.APIError
	if stateID != "" {
		state, err = s.Store.Get(stateID)
		if err != nil {
			return
		}
		stateError = state.Error
	}

	err = s.Graph.DryRun(func(ctx *newinteraction.Context) (_ *newinteraction.Graph, err error) {
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

	newStateID := intent.StateID
	if newStateID == "" {
		newStateID = NewID()
	}
	state = &State{
		ID:              newStateID,
		RedirectURI:     intent.RedirectURI,
		KeepState:       intent.KeepState,
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

	err = s.Graph.DryRun(func(ctx *newinteraction.Context) (_ *newinteraction.Graph, err error) {
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
	err = s.Graph.DryRun(func(ctx *newinteraction.Context) (newGraph *newinteraction.Graph, err error) {
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

	err = s.Graph.DryRun(func(ctx *newinteraction.Context) (newGraph *newinteraction.Graph, err error) {
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

func (s *Service) afterPost(state *State, graph *newinteraction.Graph, edges []newinteraction.Edge, inputError error) (*Result, error) {
	finished := graph != nil && len(edges) == 0 && inputError == nil
	// The graph finished. Apply its effect permanently
	if finished {
		s.Logger.Debugf("afterPost run graph")
		inputError = s.Graph.Run(graph, state.KeepState)
	}

	state.Error = skyerr.AsAPIError(inputError)
	if graph != nil {
		state.GraphInstanceID = graph.InstanceID
	}

	err := s.Store.Create(state)
	if err != nil {
		return nil, err
	}

	// Case: error
	if state.Error != nil {
		s.Logger.WithFields(map[string]interface{}{
			"error": state.Error,
		}).Debugf("afterPost error")
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

		return &Result{
			state:       state,
			redirectURI: state.RedirectURI,
			cookies:     cookies,
		}, nil
	}

	s.Logger.Debugf("afterPost transition")

	node := graph.CurrentNode()
	firstEdge := edges[0]

	// Case: the node has redirect URI.
	if a, ok := node.(RedirectURIGetter); ok {
		s.Logger.Debugf("afterPost transition to node redirect URI")
		return &Result{
			state:       state,
			redirectURI: a.GetRedirectURI(),
		}, nil
	}

	var path string
	// Case: transition
	switch node.(type) {
	case *nodes.NodeAuthenticationBegin:
		switch firstEdge.(type) {
		case *nodes.EdgeAuthenticationPassword:
			path = "/enter_password"
		default:
			panic(fmt.Errorf("webapp: unexpected edge: %T", firstEdge))
		}
	case *nodes.NodeAuthenticationOOBTrigger:
		path = "/oob_otp"
	default:
		panic(fmt.Errorf("webapp: unexpected node: %T", node))
	}

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
