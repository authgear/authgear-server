package newinteraction

import (
	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

var ErrStateNotFound = errorutil.New("invalid state or state not found")

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

func (s *Service) create(graph *Graph) error {
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

func (s *Service) NewGraph(ctx *Context, intent Intent) (*Graph, error) {
	graph := newGraph(intent)
	node, err := graph.Intent.InstantiateRootNode(ctx, graph)
	if err != nil {
		return nil, err
	}

	graph = graph.appendingNode(node)
	err = node.Prepare(ctx, graph)
	if err != nil {
		return nil, err
	}
	err = node.Apply(ctx.perform, graph)
	if err != nil {
		return nil, err
	}

	return graph, nil
}

func (s *Service) Get(instanceID string) (*Graph, error) {
	return s.Store.GetGraphInstance(instanceID)
}

func (s *Service) DryRun(webStateID string, fn func(*Context) (*Graph, error)) (err error) {
	ctx, err := s.Context.initialize()
	if err != nil {
		return
	}

	defer func() {
		rbErr := ctx.rollback()
		if rbErr != nil {
			s.Logger.WithError(rbErr).Error("cannot rollback")
			err = errorutil.WithSecondaryError(err, rbErr)
		}
	}()

	ctx.IsDryRun = true
	ctx.WebStateID = webStateID
	graph, err := fn(ctx)
	if err != nil {
		return
	}

	if graph != nil {
		err = s.create(graph)
		return
	}

	return
}

func (s *Service) Run(webStateID string, graph *Graph, preserveGraph bool) (err error) {
	ctx, err := s.Context.initialize()
	if err != nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			rbErr := ctx.rollback()
			if rbErr != nil {
				s.Logger.WithError(rbErr).Error("cannot rollback")
			}
			panic(r)
		} else if err == nil {
			err = ctx.commit()
		} else {
			rbErr := ctx.rollback()
			if rbErr != nil {
				s.Logger.WithError(rbErr).Error("cannot rollback")
				err = errorutil.WithSecondaryError(err, rbErr)
			}
		}
	}()

	ctx.IsDryRun = false
	ctx.WebStateID = webStateID
	err = graph.Apply(ctx)
	if err != nil {
		return
	}

	if !preserveGraph {
		delErr := s.Store.DeleteGraph(graph)
		if delErr != nil {
			s.Logger.WithError(delErr).Error("cannot delete graph")
		}
	}

	return
}
