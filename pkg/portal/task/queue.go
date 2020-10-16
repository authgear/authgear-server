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

func (q *InProcessQueue) Enqueue(param task.Param) {
	q.Executor.Run(context.Background(), param)
}
