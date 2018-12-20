package async

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

func (m *MockQueue) Enqueue(name string, param interface{}, response chan error) {
	m.TasksName = append(m.TasksName, name)
	m.TasksParam = append(m.TasksParam, param)
}
