package welcemail

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
)

type MockSender struct {
	LastEmail       string
	LastUserProfile userprofile.UserProfile
}

func NewMockSender() *MockSender {
	return &MockSender{}
}

func (m *MockSender) Send(email string, userProfile userprofile.UserProfile) error {
	m.LastEmail = email
	m.LastUserProfile = userProfile
	return nil
}
