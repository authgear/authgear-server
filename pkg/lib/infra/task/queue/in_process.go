package queue

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
)

type Executor interface {
	Run(ctx *task.Context, param task.Param)
}

type InProcessQueue struct {
	Database       *appdb.Handle
	CaptureContext task.CaptureTaskContext
	Executor       Executor

	pendingTasks []task.Param `wire:"-"`
	hooked       bool         `wire:"-"`
}

func (s *InProcessQueue) Enqueue(param task.Param) {
	if s.Database != nil {
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

func (s *InProcessQueue) WillCommitTx(ctx context.Context) error {
	return nil
}

func (s *InProcessQueue) DidCommitTx(ctx context.Context) {
	// To avoid running the tasks multiple times
	// reset s.pendingTasks when we start processing the tasks
	pendingTasks := s.pendingTasks
	s.pendingTasks = nil

	for _, param := range pendingTasks {
		s.run(param)
	}
}

func (s *InProcessQueue) run(param task.Param) {
	s.Executor.Run(s.CaptureContext(), param)
}
