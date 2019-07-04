package userverify

import (
	"errors"
	"strings"
)

type MockStore struct {
	CodeByID map[string]VerifyCode
}

func (m *MockStore) CreateVerifyCode(code *VerifyCode) error {
	m.CodeByID[code.ID] = *code
	return nil
}

func (m *MockStore) MarkConsumed(codeID string) error {
	code, found := m.CodeByID[codeID]
	if !found {
		return errors.New("code not found")
	}

	code.Consumed = true
	m.CodeByID[codeID] = code

	return nil
}

func (m *MockStore) GetVerifyCodeByCode(userID string, code string) (VerifyCode, error) {
	for _, c := range m.CodeByID {
		if strings.ToLower(c.Code) == strings.ToLower(code) {
			return c, nil
		}
	}

	return VerifyCode{}, errors.New("code not found")
}

var _ Store = &MockStore{}
