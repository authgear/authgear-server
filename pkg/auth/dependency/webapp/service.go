package webapp

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/skyerr"
)

//go:generate mockgen -source=service.go -destination=service_mock_test.go -package webapp

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

type Service struct {
	Store Store
	Graph GraphService
}

func (s *Service) GetIntent(intent *Intent, stateID string) (state *State, graph *newinteraction.Graph, edges []newinteraction.Edge, err error) {
	if stateID != "" {
		return s.Get(stateID)
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

	// FIXME(webapp): Interpret graph and edges to derive redirectURI.
	// FIXME(webapp): Check if current node implements RedirectURIGetter.
	return &Result{
		state:       state,
		redirectURI: "",
	}, nil
}
