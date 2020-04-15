package async

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type Queue interface {
	Enqueue(spec TaskSpec)
	WillCommitTx() error
	DidCommitTx()
}

type queue struct {
	context     context.Context
	txContext   db.TxContext
	taskContext TaskContext

	pendingTasks []TaskSpec
	hooked       bool
	taskExecutor *Executor
}

func NewQueue(
	ctx context.Context,
	txContext db.TxContext,
	requestID string,
	tenantConfig config.TenantConfiguration,
	taskExecutor *Executor,
) Queue {
	return &queue{
		context:   ctx,
		txContext: txContext,
		taskContext: TaskContext{
			RequestID:    requestID,
			TenantConfig: tenantConfig,
		},
		taskExecutor: taskExecutor,
	}
}

func (s *queue) Enqueue(spec TaskSpec) {
	if s.txContext != nil && s.txContext.HasTx() {
		s.pendingTasks = append(s.pendingTasks, spec)
		if !s.hooked {
			s.txContext.UseHook(s)
			s.hooked = true
		}
	} else {
		// No transaction context -> execute immediately.
		s.taskExecutor.Execute(s.taskContext, spec.Name, spec.Param)
	}
}

func (s *queue) WillCommitTx() error {
	return nil
}

func (s *queue) DidCommitTx() {
	for _, task := range s.pendingTasks {
		s.taskExecutor.Execute(s.taskContext, task.Name, task.Param)
	}
	s.pendingTasks = nil
}
