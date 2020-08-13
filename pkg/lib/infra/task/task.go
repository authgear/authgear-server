package task

import (
	"context"
)

type Spec struct {
	Name  string
	Param interface{}
}

type Task interface {
	Run(context context.Context, param interface{}) error
}

type Registry interface {
	Register(name string, task Task)
}

type Queue interface {
	Enqueue(spec Spec)
}
