package newinteraction

import (
	"github.com/authgear/authgear-server/pkg/core/errors"
	"github.com/authgear/authgear-server/pkg/log"
)

type Logger struct{ *log.Logger }

type Service struct {
	Logger  Logger
	Context *Context
	Store   *Store
}

func (s *Service) Create(graph *Graph) error {
	if graph.InstanceID != "" {
		panic("interaction: cannot create an existing graph")
	}
	graph.InstanceID = newInstanceID()
	return s.Store.Create(graph)
}

func (s *Service) Delete(instanceID string) error {
	return s.Store.Delete(instanceID)
}

func (s *Service) Get(instanceID string) (*Graph, error) {
	return s.Store.Get(instanceID)
}

func (s *Service) WithContext(fn func(*Context) (*Graph, error)) error {
	ctx, err := s.Context.initialize()
	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			err = ctx.commit()
		} else {
			rbErr := ctx.rollback()
			if rbErr != nil {
				s.Logger.WithError(rbErr).Error("cannot rollback")
				err = errors.WithSecondaryError(err, rbErr)
			}
		}
	}()

	graph, err := fn(ctx)
	if err != nil {
		return err
	}

	// Do not create graph if no graph is returned
	if graph == nil {
		return nil
	}
	return s.Create(graph)
}

func (s *Service) NewGraph(ctx *Context, intent Intent) (*Graph, error) {
	graph := newGraph(intent)
	node, err := graph.Intent.InstantiateRootNode(ctx, graph)
	if err != nil {
		return nil, err
	}

	graph = graph.appendingNode(node)
	err = node.Apply(ctx, graph)
	if err != nil {
		return nil, err
	}

	return graph, nil
}
