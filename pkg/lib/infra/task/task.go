package task

import (
	"context"
)

type Param interface {
	TaskName() string
}

type Task interface {
	Run(context context.Context, param Param) error
}

type Registry interface {
	Register(name string, task Task)
}

type Queue interface {
	Enqueue(taskParam Param)
}
