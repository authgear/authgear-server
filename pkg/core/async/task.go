package async

import (
	"context"
)

type TaskSpec struct {
	Name  string
	Param interface{}
}

type Task interface {
	Run(context context.Context, param interface{}) error
}
