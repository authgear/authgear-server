package interaction

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

var ErrGraphNotFound = errors.New("invalid graph or graph not found")

type Store interface {
	CreateGraph(ctx context.Context, graph *Graph) error
	CreateGraphInstance(ctx context.Context, graph *Graph) error
	GetGraphInstance(ctx context.Context, instanceID string) (*Graph, error)
	DeleteGraph(ctx context.Context, graph *Graph) error
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("interaction")} }

type Service struct {
	Logger  Logger
	Context *Context
	Store   Store
}

func (s *Service) create(ctx context.Context, graph *Graph) error {
	if graph.InstanceID != "" {
		panic("interaction: cannot re-create an existing graph instance")
	}
	graph.InstanceID = newInstanceID()

	if graph.GraphID == "" {
		graph.GraphID = newGraphID()
		return s.Store.CreateGraph(ctx, graph)
	}
	return s.Store.CreateGraphInstance(ctx, graph)
}

func (s *Service) NewGraph(ctx context.Context, interactionCtx *Context, intent Intent) (*Graph, error) {
	graph := newGraph(intent)
	node, err := graph.Intent.InstantiateRootNode(ctx, interactionCtx, graph)
	if err != nil {
		return nil, err
	}

	graph = graph.appendingNode(node)
	err = node.Prepare(ctx, interactionCtx, graph)
	if err != nil {
		return nil, err
	}
	effs, err := node.GetEffects(ctx)
	if err != nil {
		return nil, err
	}
	for _, eff := range effs {
		err = eff.apply(ctx, interactionCtx, graph, 0)
		if err != nil {
			return nil, err
		}
	}

	return graph, nil
}

func (s *Service) Get(ctx context.Context, instanceID string) (*Graph, error) {
	return s.Store.GetGraphInstance(ctx, instanceID)
}

func (s *Service) DryRun(ctx context.Context, contextValues ContextValues, fn func(ctx context.Context, interactionCtx *Context) (*Graph, error)) (err error) {
	interactionCtx, err := s.Context.initialize(ctx)
	if err != nil {
		return
	}

	defer func() {
		rbErr := interactionCtx.rollback(ctx)
		if rbErr != nil {
			s.Logger.WithError(rbErr).Error("cannot rollback")
			err = errorutil.WithSecondaryError(err, rbErr)
		}
	}()

	interactionCtx.IsCommitting = false
	interactionCtx.WebSessionID = contextValues.WebSessionID
	interactionCtx.OAuthSessionID = contextValues.OAuthSessionID
	graph, err := fn(ctx, interactionCtx)
	if err != nil {
		return
	}

	if graph != nil {
		err = s.create(ctx, graph)
		return
	}

	return
}

func (s *Service) Run(ctx context.Context, contextValues ContextValues, graph *Graph) (err error) {
	interactionCtx, err := s.Context.initialize(ctx)
	if err != nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			rbErr := interactionCtx.rollback(ctx)
			if rbErr != nil {
				s.Logger.WithError(rbErr).Error("cannot rollback")
			}
			panic(r)
		} else if err == nil {
			err = interactionCtx.commit(ctx)
		} else {
			rbErr := interactionCtx.rollback(ctx)
			if rbErr != nil {
				s.Logger.WithError(rbErr).Error("cannot rollback")
				err = errorutil.WithSecondaryError(err, rbErr)
			}
		}
	}()

	interactionCtx.IsCommitting = false
	interactionCtx.WebSessionID = contextValues.WebSessionID
	interactionCtx.OAuthSessionID = contextValues.OAuthSessionID
	err = graph.Apply(ctx, interactionCtx)
	if err != nil {
		return
	}

	interactionCtx.IsCommitting = true
	err = graph.Apply(ctx, interactionCtx)
	if err != nil {
		return
	}

	// TODO(interaction): Side effects is already committed, how to handle the subsequent errors?
	err = s.Store.DeleteGraph(ctx, graph)
	if err != nil {
		return
	}

	return
}

func (s *Service) Accept(goCtx context.Context, ctx *Context, graph *Graph, input interface{}) (*Graph, []Edge, error) {
	return graph.accept(goCtx, ctx, input)
}
