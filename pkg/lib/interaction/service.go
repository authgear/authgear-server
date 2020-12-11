package interaction

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

var ErrGraphNotFound = errors.New("invalid graph or graph not found")

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
	effs, err := node.GetEffects()
	if err != nil {
		return nil, err
	}
	for _, eff := range effs {
		err = eff.apply(ctx, graph, 0)
		if err != nil {
			return nil, err
		}
	}

	return graph, nil
}

func (s *Service) Get(instanceID string) (*Graph, error) {
	return s.Store.GetGraphInstance(instanceID)
}

func (s *Service) DryRun(webSessionID string, fn func(*Context) (*Graph, error)) (err error) {
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

	ctx.IsCommitting = false
	ctx.WebSessionID = webSessionID
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

func (s *Service) Run(webSessionID string, graph *Graph) (err error) {
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

	ctx.IsCommitting = false
	ctx.WebSessionID = webSessionID
	err = graph.Apply(ctx)
	if err != nil {
		return
	}

	ctx.IsCommitting = true
	err = graph.Apply(ctx)
	if err != nil {
		return
	}

	// TODO(interaction): Side effects is already committed, how to handle the subsequent errors?
	err = s.Store.DeleteGraph(graph)
	if err != nil {
		return
	}

	return
}

func (s *Service) Accept(ctx *Context, graph *Graph, input interface{}) (*Graph, []Edge, error) {
	bypassRateLimit := false
	var bypassInput interface{ BypassInteractionIPRateLimit() bool }
	if Input(input, &bypassInput) {
		bypassRateLimit = bypassInput.BypassInteractionIPRateLimit()
	}

	if !bypassRateLimit {
		ip := httputil.GetIP(ctx.Request, bool(ctx.TrustProxy))
		err := ctx.RateLimiter.TakeToken(RequestRateLimitBucket(ip))
		if err != nil {
			return nil, nil, err
		}
	}

	return graph.accept(ctx, input)
}
