package backgroundjob

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/authgear/authgear-server/pkg/util/panicutil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

const DefaultAfterDuration = 5 * time.Minute

type Runnable interface {
	Run(ctx context.Context) error
}

type RunnableFactory func() Runnable

var RunnerLogger = slogutil.NewLogger("backgroundjob-runner")

type Runner struct {
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

func NewRunner(ctx context.Context, runnableFactory RunnableFactory, opts ...RunnerOption) *Runner {
	runner := &Runner{
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

func (r *Runner) Start(ctx context.Context) {
	logger := RunnerLogger.GetLogger(ctx)
	r.runRunnable(ctx)
loop:
	for {
		select {
		case <-time.After(r.afterDuration):
			r.runRunnable(ctx)
		case <-r.shutdown:
			logger.Info(ctx, "shutdown gracefully")
			break loop
		case <-r.shutdownCtx.Done():
			logger.Info(ctx, "context timeout")
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

func (r *Runner) runRunnable(ctx context.Context) {
	logger := RunnerLogger.GetLogger(ctx)
	defer func() {
		if anyValue := recover(); anyValue != nil {
			err := panicutil.MakeError(anyValue)
			logger.WithError(err).Error(ctx, "panic occurred")
		}
	}()

	runner := r.runnableFactory()
	logger.Info(ctx, "start running", slog.String("runner_name", fmt.Sprintf("%T", runner)))

	err := runner.Run(r.shutdownCtx)
	if err != nil {
		logger.WithError(err).Error(ctx, "runnable ended with error")
	}
}
