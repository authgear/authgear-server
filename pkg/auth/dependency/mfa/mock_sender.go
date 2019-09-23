package mfa

type MockSender struct{}

func NewMockSender() Sender {
	return &MockSender{}
}

func (s *MockSender) Send(code string, phone string, email string) error {
	return nil
}

var (
	_ Sender = &MockSender{}
)
