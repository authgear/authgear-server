package welcemail

import "github.com/skygeario/skygear-server/pkg/auth/response"

type MockSender struct {
	LastEmail      string
	LastUserObject response.AuthResponse
}

func NewMockSender() *MockSender {
	return &MockSender{}
}

func (m *MockSender) Send(email string, userObject response.AuthResponse) error {
	m.LastEmail = email
	m.LastUserObject = userObject
	return nil
}
