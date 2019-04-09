package welcemail

import "github.com/skygeario/skygear-server/pkg/auth/response"

type MockSender struct {
	LastEmail      string
	LastUserObject response.User
}

func NewMockSender() *MockSender {
	return &MockSender{}
}

func (m *MockSender) Send(email string, user response.User) error {
	m.LastEmail = email
	m.LastUserObject = user
	return nil
}
