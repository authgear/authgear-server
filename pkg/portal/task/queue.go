package task

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/infra/task"
)

type Executor interface {
	Run(ctx context.Context, param task.Param)
}

type InProcessQueue struct {
	Executor Executor
}

func (q *InProcessQueue) Enqueue(ctx context.Context, param task.Param) {
	// Detach the deadline so that the context is not canceled along with the request.
	ctx = context.WithoutCancel(ctx)
	q.Executor.Run(ctx, param)
}
