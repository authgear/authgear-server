package welcemail

import "github.com/skygeario/skygear-server/pkg/auth/model"

type MockSender struct {
	LastEmail      string
	LastUserObject model.User
}

func NewMockSender() *MockSender {
	return &MockSender{}
}

func (m *MockSender) Send(email string, user model.User) error {
	m.LastEmail = email
	m.LastUserObject = user
	return nil
}
