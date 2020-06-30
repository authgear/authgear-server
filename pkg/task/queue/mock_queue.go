package queue

import (
	"github.com/authgear/authgear-server/pkg/task"
)

type MockQueue struct {
	TasksName  []string
	TasksParam []interface{}
}

func NewMockQueue() *MockQueue {
	return &MockQueue{
		TasksName:  []string{},
		TasksParam: []interface{}{},
	}
}

func (m *MockQueue) Enqueue(spec task.Spec) {
	m.TasksName = append(m.TasksName, spec.Name)
	m.TasksParam = append(m.TasksParam, spec.Param)
}
