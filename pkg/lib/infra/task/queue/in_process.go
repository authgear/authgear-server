package queue

import (
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
)

type Executor interface {
	Run(ctx *task.Context, spec task.Spec)
}

type InProcessQueue struct {
	Database       *db.Handle
	CaptureContext task.CaptureTaskContext
	Executor       Executor

	pendingTasks []task.Spec `wire:"-"`
	hooked       bool        `wire:"-"`
}

func (s *InProcessQueue) Enqueue(param task.Param) {
	spec := task.Spec{
		Name:  param.TaskName(),
		Param: param,
	}
	if s.Database != nil && s.Database.HasTx() {
		s.pendingTasks = append(s.pendingTasks, spec)
		if !s.hooked {
			s.Database.UseHook(s)
			s.hooked = true
		}
	} else {
		// No transaction context -> run immediately.
		s.run(spec)
	}
}

func (s *InProcessQueue) WillCommitTx() error {
	return nil
}

func (s *InProcessQueue) DidCommitTx() {
	for _, task := range s.pendingTasks {
		s.run(task)
	}
	s.pendingTasks = nil
}

func (s *InProcessQueue) run(spec task.Spec) {
	s.Executor.Run(s.CaptureContext(), spec)
}
