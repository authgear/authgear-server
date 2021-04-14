package queue

import (
	db "github.com/authgear/authgear-server/pkg/lib/infra/db/tenant"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
)

type Executor interface {
	Run(ctx *task.Context, param task.Param)
}

type InProcessQueue struct {
	Database       *db.Handle
	CaptureContext task.CaptureTaskContext
	Executor       Executor

	pendingTasks []task.Param `wire:"-"`
	hooked       bool         `wire:"-"`
}

func (s *InProcessQueue) Enqueue(param task.Param) {
	if s.Database != nil && s.Database.HasTx() {
		s.pendingTasks = append(s.pendingTasks, param)
		if !s.hooked {
			s.Database.UseHook(s)
			s.hooked = true
		}
	} else {
		// No transaction context -> run immediately.
		s.run(param)
	}
}

func (s *InProcessQueue) WillCommitTx() error {
	return nil
}

func (s *InProcessQueue) DidCommitTx() {
	for _, param := range s.pendingTasks {
		s.run(param)
	}
	s.pendingTasks = nil
}

func (s *InProcessQueue) run(param task.Param) {
	s.Executor.Run(s.CaptureContext(), param)
}
