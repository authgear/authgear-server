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

func (s *Service) WithContext(fn func(*Context) error) error {
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

	err = fn(ctx)
	return err
}

func (s *Service) NewGraph(ctx *Context, intent Intent) (*Graph, error) {
	graph := newGraph(intent)
	node, err := graph.Intent.DeriveFirstNode(ctx)
	if err != nil {
		return nil, err
	}

	graph = graph.AppendingNode(node)
	if err != nil {
		return nil, err
	}

	return graph, nil
}
