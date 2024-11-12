package backgroundjob

import (
	"context"
	"time"

	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/panicutil"
)

const DefaultAfterDuration = 5 * time.Minute

type Runnable interface {
	Run(ctx context.Context) error
}

type RunnableFactory func() Runnable

type Runner struct {
	logger          *log.Logger
	runnableFactory RunnableFactory
	afterDuration   time.Duration
	// shutdown is for breaking the loop.
	shutdown chan struct{}
	// shutdown blocks Stop until the loop has ended.
	shutdownDone chan struct{}
	// shutdownCtx is for shutdown timeout.
	shutdownCtx context.Context
}

type RunnerOption interface {
	apply(runner *Runner)
}

type afterDurationOption time.Duration

func WithAfterDuration(d time.Duration) RunnerOption {
	return afterDurationOption(d)
}

func (o afterDurationOption) apply(runner *Runner) {
	runner.afterDuration = time.Duration(o)
}

func NewRunner(ctx context.Context, logger *log.Logger, runnableFactory RunnableFactory, opts ...RunnerOption) *Runner {
	runner := &Runner{
		logger:          logger,
		runnableFactory: runnableFactory,
		afterDuration:   DefaultAfterDuration,
		shutdown:        make(chan struct{}),
		shutdownDone:    make(chan struct{}),
		shutdownCtx:     ctx,
	}
	for _, opt := range opts {
		opt.apply(runner)
	}
	return runner
}

func (r *Runner) Start() {
	r.runRunnable()
loop:
	for {
		select {
		case <-time.After(r.afterDuration):
			r.runRunnable()
		case <-r.shutdown:
			r.logger.Infof("shutdown gracefully")
			break loop
		case <-r.shutdownCtx.Done():
			r.logger.Infof("context timeout")
			break loop
		}
	}
	close(r.shutdownDone)
}

func (r *Runner) Stop(ctx context.Context) {
	r.shutdownCtx = ctx
	close(r.shutdown)
	<-r.shutdownDone
}

func (r *Runner) runRunnable() {
	defer func() {
		if anyValue := recover(); anyValue != nil {
			err := panicutil.MakeError(anyValue)
			r.logger.WithError(err).Error("panic occurred")
		}
	}()

	err := r.runnableFactory().Run(r.shutdownCtx)
	if err != nil {
		r.logger.WithError(err).Errorf("runnable ended with error")
	}
}
