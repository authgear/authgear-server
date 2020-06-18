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

type Queue interface {
	Enqueue(spec Spec)
}
