package newinteraction

import (
	"github.com/authgear/authgear-server/pkg/core/errors"
	"github.com/authgear/authgear-server/pkg/log"
)

var ErrStateNotFound = errors.New("invalid state or state not found")

type Store interface {
	CreateGraph(graph *Graph) error
	CreateGraphInstance(graph *Graph) error
	GetGraphInstance(instanceID string) (*Graph, error)
	DeleteGraph(graph *Graph) error
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("interaction")} }

type Service struct {
	Logger  Logger
	Context *Context
	Store   Store
}

func (s *Service) Create(graph *Graph) error {
	if graph.InstanceID != "" {
		panic("interaction: cannot re-create an existing graph instance")
	}
	graph.InstanceID = newInstanceID()

	if graph.GraphID == "" {
		graph.GraphID = newGraphID()
		return s.Store.CreateGraph(graph)
	}
	return s.Store.CreateGraphInstance(graph)
}

func (s *Service) Delete(graph *Graph) error {
	return s.Store.DeleteGraph(graph)
}

func (s *Service) Get(instanceID string) (*Graph, error) {
	return s.Store.GetGraphInstance(instanceID)
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
