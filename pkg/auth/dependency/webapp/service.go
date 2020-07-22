package webapp

import (
	"errors"
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/skyerr"
)

//go:generate mockgen -source=service.go -destination=service_mock_test.go -package webapp

var ErrRollback = errors.New("rollback")

type ErrorRedirectURIGetter interface {
	GetErrorRedirectURI() string
}

type Store interface {
	Create(state *State) error
	Get(instanceID string) (*State, error)
}

type GraphService interface {
	Get(instanceID string) (*newinteraction.Graph, error)
	WithContext(fn func(*newinteraction.Context) (*newinteraction.Graph, error)) error
	NewGraph(ctx *newinteraction.Context, intent newinteraction.Intent) (*newinteraction.Graph, error)
}

type Service struct {
	Store Store
	Graph GraphService
}

func (s *Service) GetIntent(intent *Intent, stateID string) (state *State, graph *newinteraction.Graph, edges []newinteraction.Edge, err error) {
	if stateID != "" {
		return s.Get(stateID)
	}

	errRollback := s.Graph.WithContext(func(ctx *newinteraction.Context) (_ *newinteraction.Graph, err error) {
		graph, err = s.Graph.NewGraph(ctx, intent.Intent)
		if err != nil {
			return
		}

		node := graph.CurrentNode()
		edges, err = node.DeriveEdges(ctx, graph)
		if err != nil {
			return
		}

		err = ErrRollback
		return
	})
	if !errors.Is(errRollback, ErrRollback) {
		err = errRollback
		return
	}

	state = &State{
		ID:              newID(),
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

	errRollback := s.Graph.WithContext(func(ctx *newinteraction.Context) (_ *newinteraction.Graph, err error) {
		node := graph.CurrentNode()
		edges, err = node.DeriveEdges(ctx, graph)
		if err != nil {
			return
		}

		err = ErrRollback
		return
	})
	if !errors.Is(errRollback, ErrRollback) {
		err = errRollback
		return
	}

	return
}

func (s *Service) PostIntent(intent *Intent, inputer func() (interface{}, error)) (result *Result, err error) {
	state := &State{
		ID:          newID(),
		RedirectURI: intent.RedirectURI,
		KeepState:   intent.KeepState,
	}

	var graph *newinteraction.Graph
	var edges []newinteraction.Edge
	// FIXME(webapp): How to convey the intention that we only want to
	// commit if the interaction finishes?
	err = s.Graph.WithContext(func(ctx *newinteraction.Context) (newGraph *newinteraction.Graph, err error) {
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
	if err != nil {
		result, err = s.makeResult(state, graph, edges, err)
		if err != nil {
			return
		}
		return
	}

	result, err = s.makeResult(state, graph, edges, err)
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
	state.ID = newID()

	var edges []newinteraction.Edge
	graph, err := s.Graph.Get(state.GraphInstanceID)
	if err != nil {
		return nil, err
	}

	// FIXME(webapp): How to convey the intention that we only want to
	// commit if the interaction finishes?
	err = s.Graph.WithContext(func(ctx *newinteraction.Context) (newGraph *newinteraction.Graph, err error) {
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
	if err != nil {
		result, err = s.makeResult(state, graph, edges, err)
		if err != nil {
			return
		}
		return
	}

	result, err = s.makeResult(state, graph, edges, err)
	if err != nil {
		return
	}

	return
}

func (s *Service) makeResult(state *State, graph *newinteraction.Graph, edges []newinteraction.Edge, inputError error) (*Result, error) {
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
	if len(edges) == 0 {
		return &Result{
			state:       state,
			redirectURI: state.RedirectURI,
			// FIXME(webapp): get cookies somewhere?
		}, nil
	}

	// FIXME(webapp): Interpret graph and edges to derive redirectURI.
	return &Result{
		state:       state,
		redirectURI: "",
	}, nil
}
