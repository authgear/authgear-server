package loginid

import (
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
)

type MockLoginIDChecker struct {
	Err error
}

var _ LoginIDChecker = &MockLoginIDChecker{}

func (c *MockLoginIDChecker) ValidateOne(loginID LoginID) error {
	return c.Err
}

func (c *MockLoginIDChecker) Validate(loginIDs []LoginID) error {
	return c.Err
}

func (c *MockLoginIDChecker) CheckType(loginIDKey string, standardKey metadata.StandardKey) bool {
	return loginIDKey == string(standardKey)
}

func (c *MockLoginIDChecker) StandardKey(loginIDKey string) (metadata.StandardKey, bool) {
	return metadata.StandardKey(loginIDKey), true
}
