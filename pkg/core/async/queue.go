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
	context   context.Context
	txContext db.TxContext

	// Request-scoped fields
	requestID    string
	tenantConfig *config.TenantConfiguration

	pendingTasks []TaskSpec
	hooked       bool
	taskExecutor *Executor
}

func NewQueue(
	ctx context.Context,
	txContext db.TxContext,
	requestID string,
	tenantConfig *config.TenantConfiguration,
	taskExecutor *Executor,
) Queue {
	return &queue{
		context:      ctx,
		txContext:    txContext,
		requestID:    requestID,
		tenantConfig: tenantConfig,
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
		s.execute(spec)
	}
}

func (s *queue) WillCommitTx() error {
	return nil
}

func (s *queue) DidCommitTx() {
	for _, task := range s.pendingTasks {
		s.execute(task)
	}
	s.pendingTasks = nil
}

func (s *queue) execute(spec TaskSpec) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, s.requestID)
	ctx = config.WithTenantConfig(ctx, s.tenantConfig)
	s.taskExecutor.Execute(ctx, spec)
}
