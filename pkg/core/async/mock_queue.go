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

func (m *MockQueue) Enqueue(spec TaskSpec) {
	m.TasksName = append(m.TasksName, spec.Name)
	m.TasksParam = append(m.TasksParam, spec.Param)
}

func (m *MockQueue) WillCommitTx() error {
	return nil
}

func (m *MockQueue) DidCommitTx() {
}
